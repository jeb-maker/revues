package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// CSRFToken derives a CSRF token from the session token and secret.
func CSRFToken(sessionToken, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte("csrf:"))
	_, _ = mac.Write([]byte(sessionToken))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// ValidateCSRF checks a submitted CSRF token against the session token.
func ValidateCSRF(sessionToken, secret, submitted string) bool {
	if sessionToken == "" || submitted == "" {
		return false
	}

	expected := CSRFToken(sessionToken, secret)
	return ConstantTimeEqual(expected, submitted)
}

// SignOAuthPayload returns an HMAC for OAuth pending cookie data.
func SignOAuthPayload(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte("oauth:"))
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// VerifyOAuthPayload validates payload signature.
func VerifyOAuthPayload(payload, secret, signature string) bool {
	expected := SignOAuthPayload(payload, secret)
	return ConstantTimeEqual(expected, signature)
}
