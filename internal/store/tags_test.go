package store_test

import (
	"context"
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/store"
)

func TestNormalizeTags(t *testing.T) {
	got := store.NormalizeTags([]string{" K8s ", "SECU", "k8s", ""})
	want := []string{"k8s", "secu"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("NormalizeTags() = %v, want %v", got, want)
	}
}

func TestTemplateMatchesSubject_Intersection(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = defaultOrgCtx(ctx, st)

	lead, err := st.UpsertGitHubUser(ctx, 1, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	if err != nil {
		t.Fatalf("UpsertGitHubUser(): %v", err)
	}
	subject, err := st.CreateSubject(ctx, "P", "", lead.ID, []string{"k8s", "frontend"})
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}

	globalTpl, _, err := st.CreateChecklistTemplate(ctx, "Global", lead.ID, nil, []store.TemplateItemInput{{Label: "A"}})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(global): %v", err)
	}
	matchTpl, _, err := st.CreateChecklistTemplate(ctx, "K8s", lead.ID, []string{"k8s", "secu"}, []store.TemplateItemInput{{Label: "B"}})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(match): %v", err)
	}
	otherTpl, _, err := st.CreateChecklistTemplate(ctx, "Other", lead.ID, []string{"mobile"}, []store.TemplateItemInput{{Label: "C"}})
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(other): %v", err)
	}

	assertMatch := func(templateID int64, want bool) {
		t.Helper()
		ok, matchErr := st.TemplateMatchesSubject(ctx, subject.ID, templateID)
		if matchErr != nil {
			t.Fatalf("TemplateMatchesSubject(%d): %v", templateID, matchErr)
		}
		if ok != want {
			t.Fatalf("TemplateMatchesSubject(%d) = %v, want %v", templateID, ok, want)
		}
	}

	assertMatch(globalTpl.ID, true)
	assertMatch(matchTpl.ID, true)
	assertMatch(otherTpl.ID, false)

	list, err := st.ListChecklistTemplates(ctx, subject.ID)
	if err != nil {
		t.Fatalf("ListChecklistTemplates(): %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len(list) = %d, want 2", len(list))
	}
}

func TestTemplateMatchesSubject_SubjectWithoutDomains(t *testing.T) {
	ctx := context.Background()
	db := openMemoryDB(t)
	st := store.New(db)
	ctx = defaultOrgCtx(ctx, st)

	lead, _ := st.UpsertGitHubUser(ctx, 2, "lead", "lead@example.com", "Lead", "", auth.RoleEditor)
	subject, err := st.CreateSubject(ctx, "P", "", lead.ID, nil)
	if err != nil {
		t.Fatalf("CreateSubject(): %v", err)
	}

	globalTpl, _, err := st.CreateChecklistTemplate(ctx, "Global", lead.ID, nil, nil)
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(): %v", err)
	}
	taggedTpl, _, err := st.CreateChecklistTemplate(ctx, "Tagged", lead.ID, []string{"k8s"}, nil)
	if err != nil {
		t.Fatalf("CreateChecklistTemplate(tagged): %v", err)
	}

	ok, err := st.TemplateMatchesSubject(ctx, subject.ID, globalTpl.ID)
	if err != nil || !ok {
		t.Fatalf("global match = %v, %v", ok, err)
	}
	ok, err = st.TemplateMatchesSubject(ctx, subject.ID, taggedTpl.ID)
	if err != nil || ok {
		t.Fatalf("tagged match = %v, %v; want false", ok, err)
	}
}
