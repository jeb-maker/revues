package crypto_test

import (
	"encoding/base64"
	"testing"

	"github.com/jeb-maker/revues/internal/crypto"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := make([]byte, crypto.KeySize)
	for i := range key {
		key[i] = byte(i)
	}

	plaintext := []byte(`{"host":"smtp.example.com","port":587}`)
	ciphertext, err := crypto.Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt(): %v", err)
	}

	got, err := crypto.Decrypt(key, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt(): %v", err)
	}
	if string(got) != string(plaintext) {
		t.Fatalf("Decrypt() = %q, want %q", got, plaintext)
	}
}

func TestDecodeKey(t *testing.T) {
	key := make([]byte, crypto.KeySize)
	encoded := base64.StdEncoding.EncodeToString(key)

	got, err := crypto.DecodeKey(encoded)
	if err != nil {
		t.Fatalf("DecodeKey(): %v", err)
	}
	if len(got) != crypto.KeySize {
		t.Fatalf("key length = %d, want %d", len(got), crypto.KeySize)
	}
}

func TestDecodeKeyInvalid(t *testing.T) {
	if _, err := crypto.DecodeKey(""); err == nil {
		t.Fatal("expected error for empty key")
	}
	if _, err := crypto.DecodeKey(base64.StdEncoding.EncodeToString([]byte("short"))); err == nil {
		t.Fatal("expected error for short key")
	}
}
