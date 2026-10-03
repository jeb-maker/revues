package testutil

import (
	"encoding/base64"

	"github.com/jeb-maker/revues/internal/crypto"
)

// EncryptionKey returns a deterministic, valid base64 AES-256 key for tests.
func EncryptionKey() string {
	key := make([]byte, crypto.KeySize)
	for i := range key {
		key[i] = byte(i + 1)
	}
	return base64.StdEncoding.EncodeToString(key)
}
