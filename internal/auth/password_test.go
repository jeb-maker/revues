package auth

import (
	"strings"
	"testing"
)

func TestHashPassword_RoundTrip(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("correct-horse-battery")
	if err != nil {
		t.Fatalf("HashPassword(): %v", err)
	}
	if !strings.HasPrefix(hash, "argon2id$") {
		t.Fatalf("unexpected hash encoding: %q", hash)
	}
	if !VerifyPassword(hash, "correct-horse-battery") {
		t.Fatal("VerifyPassword() rejected correct password")
	}
	if VerifyPassword(hash, "wrong-password") {
		t.Fatal("VerifyPassword() accepted wrong password")
	}
	if VerifyPassword("", "correct-horse-battery") {
		t.Fatal("VerifyPassword() accepted empty hash")
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	t.Parallel()

	if err := ValidatePasswordStrength("short"); err == nil {
		t.Fatal("expected short password error")
	}
	if err := ValidatePasswordStrength(strings.Repeat("a", 8)); err != nil {
		t.Fatalf("8 chars should pass: %v", err)
	}
	if err := ValidatePasswordStrength(strings.Repeat("a", 73)); err == nil {
		t.Fatal("expected long password error")
	}
}
