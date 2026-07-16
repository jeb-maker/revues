// Package main seeds the development database with demo subjects data.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/config"
	"github.com/jeb-maker/revues/internal/orgctx"
	"github.com/jeb-maker/revues/internal/store"
)

const (
	seedMarkerV1       = "demo_seed_v1"
	seedMarkerV2       = "demo_seed_v2"
	seedMarkerV3       = "demo_seed_v3"
	seedMarkerV4       = "demo_seed_v4"
	targetSubjectCount = 100
	bulkTemplateCount  = 25
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	db, err := store.Open(ctx, cfg.DatabasePath, cfg.DBMaxOpenConns)
	if err != nil {
		slog.Error("database open failed", "err", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("database close failed", "err", closeErr)
		}
	}()

	if migrateErr := store.Migrate(ctx, db); migrateErr != nil {
		slog.Error("database migrate failed", "err", migrateErr)
		os.Exit(1)
	}

	st := store.New(db)

	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		slog.Error("default organization", "err", err)
		os.Exit(1)
	}
	ctx = orgctx.WithOrganizationID(ctx, defaultOrg.ID)

	admin, err := resolveAdmin(ctx, st, cfg.BootstrapAdminEmail)
	if err != nil {
		slog.Error("resolve admin", "err", err)
		os.Exit(1)
	}

	var ran int
	if !hasSeedMarker(ctx, st, seedMarkerV1) {
		if err := seedV1(ctx, st, admin); err != nil {
			slog.Error("seed v1 failed", "err", err)
			os.Exit(1)
		}
		if err := markSeeded(ctx, st, seedMarkerV1); err != nil {
			slog.Error("seed marker v1", "err", err)
			os.Exit(1)
		}
		fmt.Println("Phase v1 : utilisateurs, sujets de base, modèles et revues.")
		ran++
	}
	if !hasSeedMarker(ctx, st, seedMarkerV2) {
		n, err := seedBulkSubjects(ctx, st, admin)
		if err != nil {
			slog.Error("seed v2 failed", "err", err)
			os.Exit(1)
		}
		if err := markSeeded(ctx, st, seedMarkerV2); err != nil {
			slog.Error("seed marker v2", "err", err)
			os.Exit(1)
		}
		fmt.Printf("Phase v2 : %d sujets supplémentaires ajoutés.\n", n)
		ran++
	}
	if !hasSeedMarker(ctx, st, seedMarkerV3) {
		n, total, err := seedToSubjectCount(ctx, st, admin, targetSubjectCount)
		if err != nil {
			slog.Error("seed v3 failed", "err", err)
			os.Exit(1)
		}
		if err := markSeeded(ctx, st, seedMarkerV3); err != nil {
			slog.Error("seed marker v3", "err", err)
			os.Exit(1)
		}
		fmt.Printf("Phase v3 : %d sujets supplémentaires ajoutés (%d au total).\n", n, total)
		ran++
	}
	if !hasSeedMarker(ctx, st, seedMarkerV4) {
		templates, runs, err := seedTemplatesAndRuns(ctx, st, admin)
		if err != nil {
			slog.Error("seed v4 failed", "err", err)
			os.Exit(1)
		}
		if err := markSeeded(ctx, st, seedMarkerV4); err != nil {
			slog.Error("seed marker v4", "err", err)
			os.Exit(1)
		}
		fmt.Printf("Phase v4 : %d modèles et %d revues ajoutés.\n", templates, runs)
		ran++
	}
	if ran == 0 {
		fmt.Println("Base déjà à jour. Rien à faire.")
		return
	}
	fmt.Println("Seed terminé.")
}

func hasSeedMarker(ctx context.Context, st *store.Store, marker string) bool {
	val, err := st.GetSetting(ctx, marker)
	return err == nil && len(val) > 0
}

func markSeeded(ctx context.Context, st *store.Store, marker string) error {
	return st.UpsertSetting(ctx, marker, []byte(time.Now().UTC().Format(time.RFC3339)))
}

func resolveAdmin(ctx context.Context, st *store.Store, email string) (*store.User, error) {
	if email != "" {
		user, err := st.UserByEmail(ctx, email)
		if err == nil {
			return user, nil
		}
		if !errors.Is(err, store.ErrUserNotFound) {
			return nil, err
		}
	}

	user, err := st.UserByID(ctx, 1)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Base vide après reset : créer un admin de démo pour le seed.
	bootstrapEmail := email
	if bootstrapEmail == "" {
		bootstrapEmail = "admin@example.com"
	}
	user, err = st.UpsertGitHubUser(ctx, 1, "admin", bootstrapEmail, "Admin Revues", "", auth.RoleAdmin)
	if err != nil {
		return nil, fmt.Errorf("bootstrap admin: %w", err)
	}
	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		return nil, fmt.Errorf("default organization: %w", err)
	}
	if err := st.AddOrganizationMember(ctx, defaultOrg.ID, user.ID, store.OrgRoleOwner); err != nil {
		return nil, fmt.Errorf("bootstrap org owner: %w", err)
	}
	return user, nil
}

func seedV1(ctx context.Context, st *store.Store, admin *store.User) error {
	alice, err := st.UpsertGitHubUser(ctx, 9001, "alice", "alice@example.com", "Alice Martin", "", auth.RoleEditor)
	if err != nil {
		return fmt.Errorf("alice: %w", err)
	}
	bob, err := st.UpsertGitHubUser(ctx, 9002, "bob", "bob@example.com", "Bob Dupont", "", auth.RoleEditor)
	if err != nil {
		return fmt.Errorf("bob: %w", err)
	}
	claire, err := st.UpsertGitHubUser(ctx, 9003, "claire", "claire@example.com", "Claire Leroy", "", auth.RoleReader)
	if err != nil {
		return fmt.Errorf("claire: %w", err)
	}

	defaultOrg, err := st.OrganizationBySlug(ctx, "default")
	if err != nil {
		return err
	}
	for _, pair := range []struct {
		user *store.User
		role string
	}{
		{alice, store.OrgRoleMember},
		{bob, store.OrgRoleMember},
		{claire, store.OrgRoleMember},
	} {
		if err = st.AddOrganizationMember(ctx, defaultOrg.ID, pair.user.ID, pair.role); err != nil {
			return fmt.Errorf("org member %s: %w", pair.user.Login, err)
		}
	}

	portail, err := st.CreateSubject(ctx, "Portail client", "Application web de gestion des revues qualité.", admin.ID, []string{"web", "frontend"})
	if err != nil {
		return fmt.Errorf("subject portail: %w", err)
	}
	api, err := st.CreateSubject(ctx, "API paiements", "Microservice de traitement des transactions.", admin.ID, []string{"api", "backend"})
	if err != nil {
		return fmt.Errorf("subject api: %w", err)
	}
	if err = st.SetSubjectTags(ctx, portail.ID, []string{"prod", "critique"}); err != nil {
		return fmt.Errorf("tags portail: %w", err)
	}

	releaseTpl, _, err := st.CreateChecklistTemplate(ctx, "Revue de release", admin.ID, nil, []store.TemplateItemInput{
		{Section: "Préparation", Label: "Branche release créée et taggée", HelpText: "Tag semver sur la branche de release.", Required: true},
		{Section: "Préparation", Label: "Notes de version rédigées", Required: true},
		{Section: "Qualité", Label: "Tests unitaires verts en CI", Required: true},
		{Section: "Qualité", Label: "Tests e2e passés sur staging", Required: true},
		{Section: "Qualité", Label: "Revue de code complète", Required: true},
		{Section: "Déploiement", Label: "Migration base validée", HelpText: "Exécuter sur une copie de prod.", Required: true},
		{Section: "Déploiement", Label: "Plan de rollback documenté", Required: true},
		{Section: "Déploiement", Label: "Alertes monitoring configurées", Required: false},
	})
	if err != nil {
		return fmt.Errorf("template release: %w", err)
	}

	secTpl, _, err := st.CreateChecklistTemplate(ctx, "Revue sécurité", admin.ID, []string{"api"}, []store.TemplateItemInput{
		{Section: "Authentification", Label: "Sessions expirées correctement", Required: true},
		{Section: "Authentification", Label: "CSRF actif sur tous les POST", Required: true},
		{Section: "Données", Label: "Contrôles IDOR vérifiés", Required: true},
		{Section: "Données", Label: "Secrets hors du dépôt", Required: true},
		{Section: "Réseau", Label: "Headers de sécurité présents", Required: true},
	})
	if err != nil {
		return fmt.Errorf("template sécurité: %w", err)
	}

	dueSoon := sql.NullString{String: time.Now().UTC().Add(48 * time.Hour).Format("2006-01-02"), Valid: true}

	activeRun, err := st.CreateChecklistRun(ctx, portail.ID, releaseTpl.ID, admin.ID)
	if err != nil {
		return fmt.Errorf("active run: %w", err)
	}
	if err = st.SetRunDueDate(ctx, activeRun.ID, dueSoon); err != nil {
		return fmt.Errorf("set active run due date: %w", err)
	}
	if err = st.StartRun(ctx, activeRun.ID); err != nil {
		return fmt.Errorf("start active run: %w", err)
	}
	if err = populateActiveRun(ctx, st, activeRun.ID, admin.ID, alice.ID, bob.ID); err != nil {
		return err
	}

	// Second in-progress run (no longer created as draft).
	if _, err := st.CreateChecklistRun(ctx, api.ID, secTpl.ID, bob.ID); err != nil {
		return fmt.Errorf("second run: %w", err)
	}

	doneRun, err := st.CreateChecklistRun(ctx, portail.ID, releaseTpl.ID, admin.ID)
	if err != nil {
		return fmt.Errorf("done run: %w", err)
	}
	if err = st.StartRun(ctx, doneRun.ID); err != nil {
		return fmt.Errorf("start done run: %w", err)
	}
	doneItems, err := st.ListRunItems(ctx, doneRun.ID)
	if err != nil {
		return fmt.Errorf("done run items: %w", err)
	}
	for _, item := range doneItems {
		if err = st.UpdateRunItemStatus(ctx, doneRun.ID, item.ID, admin.ID, store.RunItemStatusOK, ""); err != nil {
			return fmt.Errorf("complete item %d: %w", item.ID, err)
		}
	}
	if err = st.CompleteRun(ctx, doneRun.ID, "Déploiement réussi, aucun incident."); err != nil {
		return fmt.Errorf("complete run: %w", err)
	}

	return nil
}

func seedBulkSubjects(ctx context.Context, st *store.Store, admin *store.User) (int, error) {
	specs := []struct {
		name        string
		description string
		domains     []string
	}{
		{"Back-office RH", "Gestion des congés, notes de frais et entretiens annuels.", []string{"rh", "interne"}},
		{"Application mobile iOS", "App native pour les équipes terrain.", []string{"mobile", "ios"}},
		{"Application mobile Android", "Parité fonctionnelle avec la version iOS.", []string{"mobile", "android"}},
		{"SIEM — logs centralisés", "Collecte et corrélation des journaux applicatifs.", []string{"ops", "sécurité"}},
		{"Data warehouse analytics", "Entrepôt de données pour les reportings métier.", []string{"data", "analytics"}},
		{"Pipeline ETL facturation", "Extraction et consolidation des flux de facturation.", []string{"data", "finance"}},
		{"Service notifications email", "Envoi transactionnel et rappels automatiques.", []string{"backend", "notifications"}},
		{"CMS marketing", "Pages vitrine et campagnes éditoriales.", []string{"web", "marketing"}},
		{"Intranet collaboratif", "Wiki, annuaire et actualités internes.", []string{"web", "interne"}},
		{"Plateforme e-learning", "Formations en ligne et suivi des parcours.", []string{"web", "formation"}},
		{"Outil de ticketing interne", "Demandes support et suivi des incidents.", []string{"support", "interne"}},
		{"Gateway API externe", "Point d'entrée unique pour les partenaires.", []string{"api", "gateway"}},
		{"Service authentification SSO", "Fédération d'identité et provisioning.", []string{"api", "auth"}},
		{"Module reporting financier", "Tableaux de bord comptables mensuels.", []string{"finance", "reporting"}},
		{"Application caisse magasin", "Encaissement et gestion des stocks magasin.", []string{"retail", "mobile"}},
		{"Système de réservation", "Prise de rendez-vous en ligne multi-sites.", []string{"web", "booking"}},
		{"Bot support client", "Assistant conversationnel de premier niveau.", []string{"support", "ia"}},
		{"Dashboard opérations", "Vue temps réel sur les indicateurs de production.", []string{"ops", "monitoring"}},
		{"Archivage documents GED", "Conservation légale et recherche full-text.", []string{"document", "compliance"}},
		{"Connecteur ERP SAP", "Synchronisation des commandes et stocks.", []string{"integration", "erp"}},
	}

	for _, spec := range specs {
		if _, err := st.CreateSubject(ctx, spec.name, spec.description, admin.ID, spec.domains); err != nil {
			return 0, fmt.Errorf("subject %q: %w", spec.name, err)
		}
	}

	return len(specs), nil
}

func seedToSubjectCount(ctx context.Context, st *store.Store, admin *store.User, target int) (added int, total int, err error) {
	existing, err := st.ListSubjects(ctx, admin.ID, true, "")
	if err != nil {
		return 0, 0, fmt.Errorf("list subjects: %w", err)
	}
	if len(existing) >= target {
		return 0, len(existing), nil
	}

	need := target - len(existing)
	prefixes := []string{"Plateforme", "Service", "Module", "Application", "Portail", "API", "Outil", "Système", "Microservice", "Dashboard"}
	domains := []string{"logistique", "finance", "marketing", "RH", "ventes", "support", "production", "qualité", "juridique", "achats", "stock", "facturation", "conformité", "innovation", "partenaires"}
	domainSets := [][]string{
		{"web", "frontend"},
		{"api", "backend"},
		{"mobile", "ios"},
		{"mobile", "android"},
		{"data", "analytics"},
		{"ops", "monitoring"},
		{"interne", "legacy"},
		{"integration", "erp"},
		nil,
	}

	for i := 0; i < need; i++ {
		n := len(existing) + i + 1
		name := fmt.Sprintf("%s %s #%03d", prefixes[i%len(prefixes)], domains[i%len(domains)], n)
		desc := fmt.Sprintf("Sujet de démonstration n°%d pour tester l'affichage à grande échelle.", n)
		domainTags := domainSets[i%len(domainSets)]

		if _, err := st.CreateSubject(ctx, name, desc, admin.ID, domainTags); err != nil {
			return added, len(existing) + added, fmt.Errorf("subject %q: %w", name, err)
		}
		added++
	}

	return added, target, nil
}

func seedTemplatesAndRuns(ctx context.Context, st *store.Store, admin *store.User) (templateCount int, runCount int, err error) {
	templateIDs, err := seedGlobalTemplates(ctx, st, admin, bulkTemplateCount)
	if err != nil {
		return 0, 0, err
	}
	templateCount = len(templateIDs)

	subjects, err := st.ListSubjects(ctx, admin.ID, true, "")
	if err != nil {
		return templateCount, 0, fmt.Errorf("list subjects: %w", err)
	}

	for i, subject := range subjects {
		existing, err := st.ListRunsBySubject(ctx, subject.ID)
		if err != nil {
			return templateCount, runCount, fmt.Errorf("list runs for subject %d: %w", subject.ID, err)
		}
		if len(existing) > 0 {
			continue
		}

		tplID := templateIDs[i%len(templateIDs)]
		var dueDate sql.NullString
		if i%3 == 0 {
			dueDate = sql.NullString{
				String: time.Now().UTC().Add(time.Duration((i%14)+1) * 24 * time.Hour).Format("2006-01-02"),
				Valid:  true,
			}
		}

		run, err := st.CreateChecklistRun(ctx, subject.ID, tplID, admin.ID)
		if err != nil {
			return templateCount, runCount, fmt.Errorf("create run for subject %d: %w", subject.ID, err)
		}
		if dueDate.Valid {
			if err := st.SetRunDueDate(ctx, run.ID, dueDate); err != nil {
				return templateCount, runCount, fmt.Errorf("set due date for run %d: %w", run.ID, err)
			}
		}

		switch i % 3 {
		case 0:
			if err := st.StartRun(ctx, run.ID); err != nil {
				return templateCount, runCount, fmt.Errorf("start run %d: %w", run.ID, err)
			}
			if err := seedPartialProgress(ctx, st, run.ID, admin.ID, i); err != nil {
				return templateCount, runCount, err
			}
		case 2:
			if err := st.StartRun(ctx, run.ID); err != nil {
				return templateCount, runCount, fmt.Errorf("start run %d: %w", run.ID, err)
			}
			if err := completeAllItems(ctx, st, run.ID, admin.ID); err != nil {
				return templateCount, runCount, err
			}
			if err := st.CompleteRun(ctx, run.ID, "Revue clôturée automatiquement (données de démo)."); err != nil {
				return templateCount, runCount, fmt.Errorf("complete run %d: %w", run.ID, err)
			}
		}

		runCount++
	}

	return templateCount, runCount, nil
}

func seedGlobalTemplates(ctx context.Context, st *store.Store, admin *store.User, count int) ([]int64, error) {
	kinds := []string{"release", "sécurité", "qualité", "déploiement", "conformité", "recette", "architecture", "performance"}
	ids := make([]int64, 0, count)

	for i := 0; i < count; i++ {
		kind := kinds[i%len(kinds)]
		name := fmt.Sprintf("Modèle %s #%02d", kind, i+1)
		tpl, _, err := st.CreateChecklistTemplate(ctx, name, admin.ID, nil, demoTemplateItems(i))
		if err != nil {
			return nil, fmt.Errorf("template %q: %w", name, err)
		}
		ids = append(ids, tpl.ID)
	}

	return ids, nil
}

func demoTemplateItems(seed int) []store.TemplateItemInput {
	sets := [][]store.TemplateItemInput{
		{
			{Section: "Préparation", Label: "Périmètre validé avec le métier", Required: true},
			{Section: "Préparation", Label: "Environnement de test disponible", Required: true},
			{Section: "Contrôles", Label: "Tests automatisés passés", Required: true},
			{Section: "Contrôles", Label: "Documentation mise à jour", Required: true},
			{Section: "Clôture", Label: "Décision go/no-go formalisée", Required: true},
		},
		{
			{Section: "Sécurité", Label: "Authentification vérifiée", Required: true},
			{Section: "Sécurité", Label: "Autorisations contrôlées", Required: true},
			{Section: "Sécurité", Label: "Données sensibles protégées", Required: true},
			{Section: "Exploitation", Label: "Logs et alertes actifs", Required: false},
		},
		{
			{Section: "Livrable", Label: "Critères d'acceptation couverts", Required: true},
			{Section: "Livrable", Label: "Régression fonctionnelle OK", Required: true},
			{Section: "Livrable", Label: "Performance acceptable", Required: false},
			{Section: "Livrable", Label: "Accessibilité vérifiée", Required: false},
		},
	}
	return sets[seed%len(sets)]
}

func seedPartialProgress(ctx context.Context, st *store.Store, runID, userID int64, seed int) error {
	items, err := st.ListRunItems(ctx, runID)
	if err != nil {
		return fmt.Errorf("list run items: %w", err)
	}
	if len(items) == 0 {
		return nil
	}

	doneUntil := len(items) / 2
	if doneUntil < 1 {
		doneUntil = 1
	}
	for i := 0; i < doneUntil && i < len(items); i++ {
		if err := st.UpdateRunItemStatus(ctx, runID, items[i].ID, userID, store.RunItemStatusOK, ""); err != nil {
			return fmt.Errorf("mark item ok: %w", err)
		}
	}
	if doneUntil < len(items) && seed%2 == 0 {
		if err := st.UpdateRunItemStatus(ctx, runID, items[doneUntil].ID, userID, store.RunItemStatusNOK, "Écart détecté lors de la revue de démo."); err != nil {
			return fmt.Errorf("mark item nok: %w", err)
		}
	}
	if seed%3 == 0 && doneUntil+1 < len(items) {
		id := userID
		if err := st.AssignRunItem(ctx, runID, items[doneUntil+1].ID, &id); err != nil {
			return fmt.Errorf("assign item: %w", err)
		}
	}

	return nil
}

func completeAllItems(ctx context.Context, st *store.Store, runID, userID int64) error {
	items, err := st.ListRunItems(ctx, runID)
	if err != nil {
		return fmt.Errorf("list run items: %w", err)
	}
	for _, item := range items {
		if err := st.UpdateRunItemStatus(ctx, runID, item.ID, userID, store.RunItemStatusOK, ""); err != nil {
			return fmt.Errorf("complete item %d: %w", item.ID, err)
		}
	}
	return nil
}

func populateActiveRun(ctx context.Context, st *store.Store, runID, adminID, aliceID, bobID int64) error {
	items, err := st.ListRunItems(ctx, runID)
	if err != nil {
		return fmt.Errorf("list run items: %w", err)
	}
	if len(items) < 8 {
		return fmt.Errorf("expected 8 items, got %d", len(items))
	}

	for i := 0; i < 3; i++ {
		if err := st.UpdateRunItemStatus(ctx, runID, items[i].ID, adminID, store.RunItemStatusOK, ""); err != nil {
			return err
		}
	}

	if err := st.UpdateRunItemStatus(ctx, runID, items[3].ID, aliceID, store.RunItemStatusNOK, "Échec sur le parcours de connexion — ticket JIRA à créer."); err != nil {
		return err
	}

	if err := st.UpdateRunItemStatus(ctx, runID, items[4].ID, bobID, store.RunItemStatusOK, ""); err != nil {
		return err
	}

	for i := 5; i < len(items); i++ {
		id := adminID
		if err := st.AssignRunItem(ctx, runID, items[i].ID, &id); err != nil {
			return fmt.Errorf("assign item %d: %w", items[i].ID, err)
		}
	}

	return nil
}
