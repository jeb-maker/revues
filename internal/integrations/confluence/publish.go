package confluence

import (
	"context"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/jeb-maker/revues/internal/store"
)

// ErrNotConfigured is returned when Confluence integration is missing or incomplete.
var ErrNotConfigured = errors.New("confluence not configured")

// ErrRunNotDone is returned when publishing a run that is not done.
var ErrRunNotDone = errors.New("run is not done")

// PublishService publishes a completed checklist run as a Confluence page.
type PublishService struct {
	Store         *store.Store
	EncryptionKey []byte
	Client        *Client
}

// PublishInput holds optional overrides for page creation.
type PublishInput struct {
	Title string
}

// PublishResult is the created (or updated) Confluence page link.
type PublishResult struct {
	URL string
}

// Publish creates a Confluence page for a done run and stores the URL on the run.
func (s *PublishService) Publish(ctx context.Context, runID int64, input PublishInput) (PublishResult, error) {
	cfg, ok, err := s.config(ctx)
	if err != nil {
		return PublishResult{}, err
	}
	if !ok || !cfg.Configured() {
		return PublishResult{}, ErrNotConfigured
	}

	run, err := s.Store.RunByID(ctx, runID)
	if err != nil {
		return PublishResult{}, err
	}
	if run.Status != store.RunStatusDone {
		return PublishResult{}, ErrRunNotDone
	}

	subject, err := s.Store.SubjectByID(ctx, run.SubjectID)
	if err != nil {
		return PublishResult{}, fmt.Errorf("load subject for confluence: %w", err)
	}

	items, err := s.Store.ListRunItems(ctx, runID)
	if err != nil {
		return PublishResult{}, fmt.Errorf("list run items for confluence: %w", err)
	}

	jiraByItem, err := s.jiraLinksForItems(ctx, items)
	if err != nil {
		return PublishResult{}, err
	}

	title, err := s.pageTitle(ctx, run, subject.Name, input.Title)
	if err != nil {
		return PublishResult{}, err
	}

	body := BuildPageHTML(PageContent{
		Title:            title,
		SubjectName:      subject.Name,
		CompletedByLogin: run.CompletedByLogin,
		CompletedAt:      completedAtString(run),
		ClosingNote:      run.ClosingNote,
		Items:            items,
		JiraByItemID:     jiraByItem,
	})

	client := s.client()
	created, err := client.CreatePage(ctx, cfg, CreatePageInput{
		Title:        title,
		SpaceKey:     cfg.SpaceKey,
		ParentPageID: cfg.ParentPageID,
		BodyHTML:     body,
	})
	if err != nil {
		return PublishResult{}, err
	}

	if err := s.Store.SetRunConfluenceURL(ctx, runID, created.URL); err != nil {
		return PublishResult{}, fmt.Errorf("store confluence url: %w", err)
	}

	return PublishResult{URL: created.URL}, nil
}

// Configured reports whether Confluence integration is stored and complete.
func (s *PublishService) Configured(ctx context.Context) (bool, error) {
	cfg, ok, err := s.config(ctx)
	if err != nil {
		return false, err
	}
	return ok && cfg.Configured(), nil
}

func (s *PublishService) pageTitle(ctx context.Context, run *store.ChecklistRun, subjectName, override string) (string, error) {
	if t := strings.TrimSpace(override); t != "" {
		return t, nil
	}
	label, err := s.Store.RunDisplayLabelForRun(ctx, run)
	if err != nil {
		return "", fmt.Errorf("run display label: %w", err)
	}
	label = strings.TrimSpace(label)
	if label == "" {
		label = "Revue"
	}
	if subjectName != "" {
		return subjectName + " — " + label, nil
	}
	return label, nil
}

func (s *PublishService) jiraLinksForItems(ctx context.Context, items []store.RunItem) (map[int64]store.IntegrationLink, error) {
	ids := make([]int64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	if len(ids) == 0 {
		return map[int64]store.IntegrationLink{}, nil
	}
	links, err := s.Store.ListIntegrationLinksByRunItemIDs(ctx, ids, store.IntegrationTypeJira)
	if err != nil {
		return nil, fmt.Errorf("list jira links for confluence: %w", err)
	}
	return links, nil
}

func (s *PublishService) config(ctx context.Context) (Config, bool, error) {
	svc := &Service{Store: s.Store, EncryptionKey: s.EncryptionKey}
	return svc.Load(ctx)
}

func (s *PublishService) client() *Client {
	if s.Client != nil {
		return s.Client
	}
	return &Client{}
}

// PageContent is the data used to render Confluence storage HTML.
type PageContent struct {
	Title            string
	SubjectName      string
	CompletedByLogin string
	CompletedAt      string
	ClosingNote      string
	Items            []store.RunItem
	JiraByItemID     map[int64]store.IntegrationLink
}

// BuildPageHTML renders a Confluence storage-format HTML body for a completed run.
func BuildPageHTML(c PageContent) string {
	var b strings.Builder
	b.WriteString("<h1>")
	b.WriteString(html.EscapeString(c.Title))
	b.WriteString("</h1>")

	b.WriteString("<p><strong>Sujet :</strong> ")
	b.WriteString(html.EscapeString(c.SubjectName))
	b.WriteString("</p>")

	if c.CompletedByLogin != "" || c.CompletedAt != "" {
		b.WriteString("<p><strong>Clôturée")
		if c.CompletedByLogin != "" {
			b.WriteString(" par</strong> @")
			b.WriteString(html.EscapeString(c.CompletedByLogin))
			if c.CompletedAt != "" {
				b.WriteString(" <strong>le</strong> ")
				b.WriteString(html.EscapeString(formatWhen(c.CompletedAt)))
			}
		} else if c.CompletedAt != "" {
			b.WriteString(" le</strong> ")
			b.WriteString(html.EscapeString(formatWhen(c.CompletedAt)))
		} else {
			b.WriteString("</strong>")
		}
		b.WriteString("</p>")
	}

	if note := strings.TrimSpace(c.ClosingNote); note != "" {
		b.WriteString("<p><strong>Note de clôture :</strong> ")
		b.WriteString(html.EscapeString(note))
		b.WriteString("</p>")
	}

	b.WriteString(`<table><tbody>`)
	b.WriteString("<tr><th>Section</th><th>Point</th><th>Statut</th><th>Commentaire</th><th>Jira</th></tr>")
	for _, item := range c.Items {
		b.WriteString("<tr><td>")
		b.WriteString(html.EscapeString(item.Section))
		b.WriteString("</td><td>")
		b.WriteString(html.EscapeString(item.Label))
		b.WriteString("</td><td>")
		b.WriteString(html.EscapeString(statusLabel(item.Status)))
		b.WriteString("</td><td>")
		b.WriteString(html.EscapeString(item.Comment))
		b.WriteString("</td><td>")
		if item.Status == store.RunItemStatusNOK {
			if link, ok := c.JiraByItemID[item.ID]; ok && link.ExternalURL != "" {
				b.WriteString(`<a href="`)
				b.WriteString(html.EscapeString(link.ExternalURL))
				b.WriteString(`">`)
				b.WriteString(html.EscapeString(link.ExternalKey))
				b.WriteString("</a>")
			}
		}
		b.WriteString("</td></tr>")
	}
	b.WriteString("</tbody></table>")
	return b.String()
}

func statusLabel(status string) string {
	switch status {
	case store.RunItemStatusOK:
		return "Validé"
	case store.RunItemStatusNOK:
		return "Non validé"
	case store.RunItemStatusNA:
		return "Non applicable"
	case store.RunItemStatusPending:
		return "En attente"
	default:
		return status
	}
}

func formatWhen(iso string) string {
	iso = strings.TrimSpace(iso)
	if iso == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return iso
	}
	return t.UTC().Format("2006-01-02 15:04 UTC")
}

func completedAtString(run *store.ChecklistRun) string {
	if run.CompletedAt.Valid {
		return run.CompletedAt.String
	}
	return ""
}
