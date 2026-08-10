package notion_test

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/integrations/notion"
)

func TestUserMessage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "nil", err: nil, want: ""},
		{name: "not configured", err: notion.ErrNotConfigured, want: "Notion n'est pas configuré"},
		{name: "database missing", err: notion.ErrDatabaseMissing, want: "Aucune base Notion par défaut"},
		{name: "already exported", err: notion.ErrAlreadyExported, want: "déjà été exportée"},
		{name: "run not done", err: notion.ErrRunNotDone, want: "revues terminées"},
		{name: "database not found", err: notion.ErrDatabaseNotFound, want: "Base Notion introuvable"},
		{
			name: "unauthorized",
			err:  &notion.APIError{Status: http.StatusUnauthorized, Code: "unauthorized", Err: notion.ErrConnectionFailed},
			want: "Jeton Notion invalide",
		},
		{
			name: "forbidden",
			err:  &notion.APIError{Status: http.StatusForbidden, Code: "restricted_resource", Err: notion.ErrConnectionFailed},
			want: "Partagez la base",
		},
		{
			name: "not found api",
			err:  &notion.APIError{Status: http.StatusNotFound, Code: "object_not_found", Err: notion.ErrExportFailed},
			want: "Ressource Notion introuvable",
		},
		{
			name: "rate limited",
			err:  &notion.APIError{Status: http.StatusTooManyRequests, Code: "rate_limited", Err: notion.ErrConnectionFailed},
			want: "limite temporairement",
		},
		{
			name: "validation",
			err:  &notion.APIError{Status: http.StatusBadRequest, Code: "validation_error", Err: notion.ErrExportFailed},
			want: "colonnes incompatibles",
		},
		{
			name: "bad request other",
			err:  &notion.APIError{Status: http.StatusBadRequest, Code: "invalid_request_url", Err: notion.ErrExportFailed},
			want: "Requête Notion invalide",
		},
		{
			name: "service unavailable",
			err:  &notion.APIError{Status: http.StatusServiceUnavailable, Err: notion.ErrConnectionFailed},
			want: "temporairement indisponible",
		},
		{
			name: "network wrapped",
			err:  fmt.Errorf("%w: dial tcp", notion.ErrConnectionFailed),
			want: "Impossible de joindre Notion",
		},
		{
			name: "export failed network",
			err:  fmt.Errorf("%w: timeout", notion.ErrExportFailed),
			want: "Impossible de joindre Notion",
		},
		{
			name: "french validation passthrough",
			err:  errors.New("sélectionnez la colonne libellé"),
			want: "sélectionnez la colonne libellé",
		},
		{
			name: "technical dump hidden",
			err:  errors.New(`notion export failed: status 500 {"object":"error"}`),
			want: "Erreur Notion. Réessayez",
		},
		{
			name: "wrapped api unauthorized",
			err:  fmt.Errorf("export: %w", &notion.APIError{Status: http.StatusUnauthorized, Err: notion.ErrExportFailed}),
			want: "Jeton Notion invalide",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := notion.UserMessage(tt.err)
			if tt.want == "" {
				if got != "" {
					t.Fatalf("UserMessage() = %q, want empty", got)
				}
				return
			}
			if !strings.Contains(got, tt.want) {
				t.Fatalf("UserMessage() = %q, want substring %q", got, tt.want)
			}
		})
	}
}
