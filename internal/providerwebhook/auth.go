package providerwebhook

import "crypto/subtle"

// ValidSharedSecret rejects missing configuration and missing credentials.
// This authenticates a shared header value, not the webhook body or its age.
func ValidSharedSecret(expected, supplied string) bool {
	if expected == "" || supplied == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(supplied)) == 1
}
