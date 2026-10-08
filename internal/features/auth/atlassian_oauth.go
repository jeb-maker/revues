package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/crypto"
	"github.com/jeb-maker/revues/internal/integrations/jira"
	"github.com/jeb-maker/revues/internal/store"
	appmiddleware "github.com/jeb-maker/revues/internal/web/middleware"
)

// DisconnectAtlassian clears stored Atlassian OAuth tokens (browser form POST + CSRF).
func (h *OAuthHandlers) DisconnectAtlassian(w http.ResponseWriter, r *http.Request) {
	user, ok := appmiddleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if err := h.Service.Store.DeleteAtlassianOAuthTokensByUserID(r.Context(), user.ID); err != nil && !errors.Is(err, store.ErrAtlassianOAuthNotFound) {
		slog.Error("disconnect atlassian", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/runs?atlassian=disconnected", http.StatusSeeOther)
}

// StartAtlassian redirects a logged-in user to Atlassian authorize (3LO).
func (h *OAuthHandlers) StartAtlassian(w http.ResponseWriter, r *http.Request) {
	if !h.Service.AtlassianOAuthConfigured() {
		http.Redirect(w, r, "/runs?atlassian="+url.QueryEscape("error"), http.StatusFound)
		return
	}

	user, ok := appmiddleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login?next="+url.QueryEscape("/runs"), http.StatusFound)
		return
	}

	state, _, err := auth.RandomToken(16)
	if err != nil {
		slog.Error("atlassian oauth state", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	payload, signature := h.Service.Sessions.BuildAtlassianOAuthCookiePayload(state, user.ID)
	h.Service.Sessions.SetOAuthCookie(w, payload, signature)

	http.Redirect(w, r, h.Service.Atlassian.AuthURL(state), http.StatusFound)
}

// CallbackAtlassian completes Atlassian 3LO and stores encrypted user tokens.
func (h *OAuthHandlers) CallbackAtlassian(w http.ResponseWriter, r *http.Request) {
	failRedirect := func(msg string) {
		http.Redirect(w, r, "/runs?atlassian="+url.QueryEscape("error")+"&reason="+url.QueryEscape(msg), http.StatusFound)
	}

	if errParam := r.URL.Query().Get("error"); errParam != "" {
		failRedirect(errParam)
		return
	}

	user, ok := appmiddleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login?next="+url.QueryEscape("/runs"), http.StatusFound)
		return
	}

	state, boundUserID, err := h.Service.Sessions.ParseAtlassianOAuthCookie(r)
	if err != nil {
		slog.Error("atlassian oauth cookie", "err", err)
		failRedirect("session oauth invalide")
		return
	}
	if boundUserID != user.ID {
		failRedirect("utilisateur oauth invalide")
		return
	}
	if !auth.ConstantTimeEqual(state, r.URL.Query().Get("state")) {
		failRedirect("state invalide")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		failRedirect("code manquant")
		return
	}

	if !h.Service.AtlassianOAuthConfigured() {
		failRedirect("oauth non configuré")
		return
	}

	token, err := h.Service.Atlassian.ExchangeCode(r.Context(), code)
	if err != nil {
		slog.Error("atlassian oauth exchange", "err", err)
		failRedirect("échec oauth")
		return
	}
	if token.RefreshToken == "" {
		failRedirect("refresh token manquant")
		return
	}

	resources, err := h.Service.Atlassian.AccessibleResources(r.Context(), token.AccessToken)
	if err != nil {
		slog.Error("atlassian accessible resources", "err", err)
		failRedirect("sites atlassian")
		return
	}

	preferBase := ""
	if encKey, keyErr := h.Service.Config.EncryptionKeyBytes(); keyErr == nil && len(encKey) == crypto.KeySize {
		if st, okStore := h.Service.Store.(*store.Store); okStore {
			jiraSvc := &jira.Service{Store: st, EncryptionKey: encKey}
			if cfg, okCfg, loadErr := jiraSvc.Load(r.Context()); loadErr == nil && okCfg {
				preferBase = cfg.BaseURL
			}
		}
	}

	resource, err := auth.PickCloudResource(resources, preferBase)
	if err != nil {
		failRedirect("aucun site jira")
		return
	}

	accountEmail, err := fetchAtlassianAccountEmail(r, h, token.AccessToken, resource.ID)
	if err != nil {
		slog.Debug("atlassian account email", "err", err)
		accountEmail = ""
	}

	encKey, err := h.Service.Config.EncryptionKeyBytes()
	if err != nil || len(encKey) != crypto.KeySize {
		failRedirect("chiffrement non configuré")
		return
	}

	accessEnc, err := crypto.Encrypt(encKey, []byte(token.AccessToken))
	if err != nil {
		slog.Error("encrypt atlassian access token", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	refreshEnc, err := crypto.Encrypt(encKey, []byte(token.RefreshToken))
	if err != nil {
		slog.Error("encrypt atlassian refresh token", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	expiresIn := token.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600
	}
	expiresAt := time.Now().UTC().Add(time.Duration(expiresIn) * time.Second).Format(time.RFC3339)

	scopes := token.Scope
	if scopes == "" {
		scopes = strings.Join(resource.Scopes, " ")
	}

	err = h.Service.Store.UpsertAtlassianOAuthTokens(r.Context(), store.AtlassianOAuthTokens{
		UserID:                user.ID,
		CloudID:               resource.ID,
		SiteURL:               strings.TrimRight(resource.URL, "/"),
		AccountEmail:          accountEmail,
		AccessTokenEncrypted:  accessEnc,
		RefreshTokenEncrypted: refreshEnc,
		ExpiresAt:             expiresAt,
		Scopes:                scopes,
	})
	if err != nil {
		slog.Error("store atlassian oauth tokens", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.Service.Sessions.ClearOAuthCookie(w)
	http.Redirect(w, r, "/runs?atlassian=connected", http.StatusFound)
}

func fetchAtlassianAccountEmail(r *http.Request, h *OAuthHandlers, accessToken, cloudID string) (string, error) {
	if cloudID == "" {
		return "", errors.New("missing cloud id")
	}
	client := &jira.Client{}
	if h.Service.Atlassian != nil && h.Service.Atlassian.HTTPClient != nil {
		client.HTTPClient = h.Service.Atlassian.HTTPClient
	}
	cfg := jira.Config{
		InstanceType: jira.InstanceCloud,
		AccessToken:  accessToken,
		CloudID:      cloudID,
		SiteURL:      "https://api.atlassian.com",
	}
	email, err := client.MyselfEmail(r.Context(), cfg)
	if err != nil {
		return "", err
	}
	return email, nil
}
