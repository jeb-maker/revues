package projects_test

import (
	"testing"

	"github.com/jeb-maker/revues/internal/auth"
	"github.com/jeb-maker/revues/internal/projects"
	"github.com/jeb-maker/revues/internal/store"
)

func TestCanCreate(t *testing.T) {
	t.Parallel()

	editor := &store.User{Role: auth.RoleEditor}
	reader := &store.User{Role: auth.RoleReader}

	if !projects.CanCreate(editor) {
		t.Fatal("editor should create projects")
	}
	if projects.CanCreate(reader) {
		t.Fatal("reader should not create projects")
	}
}

func TestCanView(t *testing.T) {
	t.Parallel()

	admin := &store.User{Role: auth.RoleAdmin}
	editor := &store.User{Role: auth.RoleEditor}

	if !projects.CanView(admin, false) {
		t.Fatal("admin should view any project")
	}
	if !projects.CanView(editor, true) {
		t.Fatal("member should view project")
	}
	if projects.CanView(editor, false) {
		t.Fatal("non-member editor should not view without membership")
	}
}
