package templates_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jeb-maker/revues/internal/web/templates"
)

func TestLayout_BreadcrumbAncestorsOnly(t *testing.T) {
	tpl, err := templates.Parse("")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	render := func(crumbs []templates.Breadcrumb) string {
		t.Helper()
		var buf bytes.Buffer
		data := templates.PageData{Title: "t", Breadcrumbs: crumbs}
		// layout_start alone needs matching end; use a minimal page define via Execute of runs_list empty is heavy.
		// Execute layout_start through a page that only uses layout — home uses ApplyPageMeta.
		if err := tpl.ExecuteTemplate(&buf, "layout_start", data); err != nil {
			t.Fatalf("layout_start: %v", err)
		}
		return buf.String()
	}

	run := templates.DefaultUILabels().Run
	root := render(templates.BCRevues(run))
	if strings.Contains(root, `Fil d'Ariane`) {
		t.Fatalf("root page should hide breadcrumb, got %s", root)
	}
	if !strings.Contains(root, `<h1 class="page-title">Revues</h1>`) {
		t.Fatalf("expected H1 Revues, got %s", root)
	}

	deep := render(templates.BCRunWizardTemplates("Alpha", 3, run))
	if !strings.Contains(deep, `Fil d'Ariane`) {
		t.Fatal("deep page should show ancestor breadcrumb")
	}
	if strings.Contains(deep, `aria-current="page"`) {
		t.Fatal("current page must not appear in breadcrumb")
	}
	if !strings.Contains(deep, `>Revues</a>`) || !strings.Contains(deep, `>Alpha</a>`) {
		t.Fatalf("expected ancestor links, got %s", deep)
	}
	if strings.Contains(deep, `Choisir un modèle`) && strings.Contains(deep, `Fil d'Ariane`) {
		// H1 has the title; breadcrumb nav must not.
		navStart := strings.Index(deep, `Fil d'Ariane`)
		navEnd := strings.Index(deep[navStart:], `</nav>`)
		nav := deep[navStart : navStart+navEnd]
		if strings.Contains(nav, `Choisir un modèle`) {
			t.Fatalf("current title leaked into breadcrumb: %s", nav)
		}
	}
}
