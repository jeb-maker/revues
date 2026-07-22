package auth

import (
	"strings"
	"testing"
)

func TestLoginErrorMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code string
		want string
	}{
		{"", ""},
		{"email non autorisé", "Connexion impossible avec ce compte GitHub"},
		{"oauth non configuré", "REVUES_GITHUB_CLIENT_ID"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			t.Parallel()
			got := LoginErrorMessage(tt.code)
			if tt.want == "" {
				if got != "" {
					t.Fatalf("LoginErrorMessage(%q) = %q, want empty", tt.code, got)
				}
				return
			}
			if !strings.Contains(got, tt.want) {
				t.Fatalf("LoginErrorMessage(%q) = %q, want substring %q", tt.code, got, tt.want)
			}
		})
	}
}
