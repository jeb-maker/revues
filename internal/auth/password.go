package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	passwordMinLen   = 8
	passwordMaxLen   = 72
	argonTime        = 1
	argonMemoryKiB   = 64 * 1024
	argonThreads     = 4
	argonKeyLen      = 32
	argonSaltLen     = 16
	argonEncodedPref = "argon2id"
)

// ValidatePasswordStrength checks length constraints for new passwords.
func ValidatePasswordStrength(password string) error {
	n := len(password)
	if n < passwordMinLen {
		return fmt.Errorf("password too short (min %d)", passwordMinLen)
	}
	if n > passwordMaxLen {
		return fmt.Errorf("password too long (max %d)", passwordMaxLen)
	}
	return nil
}

// HashPassword returns an encoded argon2id hash for storage.
func HashPassword(password string) (string, error) {
	if err := ValidatePasswordStrength(password); err != nil {
		return "", err
	}

	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("password salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemoryKiB, argonThreads, argonKeyLen)
	return fmt.Sprintf(
		"%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argonEncodedPref,
		argon2.Version,
		argonMemoryKiB,
		argonTime,
		argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// VerifyPassword checks password against an encoded argon2id hash.
func VerifyPassword(encoded, password string) bool {
	if encoded == "" || password == "" {
		return false
	}

	parts := strings.Split(encoded, "$")
	if len(parts) != 5 || parts[0] != argonEncodedPref {
		return false
	}

	var version int
	if _, err := fmt.Sscanf(parts[1], "v=%d", &version); err != nil || version != argon2.Version {
		return false
	}

	var memory uint32
	var time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[2], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	got := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}
