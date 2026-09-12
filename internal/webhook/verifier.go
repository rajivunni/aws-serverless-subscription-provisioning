package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strings"
)

type VerificationResult int

type Verifier struct {
	secret []byte
}

func NewVerifier(secret string) (*Verifier, error) {
	if secret == "" {
		return nil, errors.New("webhook verification secret cannot be empty")
	}
	return &Verifier{
		secret: []byte(secret),
	}, nil
}

func (v *Verifier) Verify(signatureHeader string, payload []byte) error {
	signature := strings.TrimSpace(signatureHeader)
	if signature == "" {
		return errors.New("signature must not be empty")
	}
	mac := hmac.New(sha256.New, v.secret)
	mac.Write(payload)

	expectedMac := mac.Sum(nil)

	expected := base64.StdEncoding.EncodeToString(expectedMac)
	sigBytes := []byte(signature)
	expBytes := []byte(expected)

	if len(sigBytes) != len(expBytes) {
		return errors.New("mismatched signature and payload length")
	}

	if subtle.ConstantTimeCompare(sigBytes, expBytes) != 1 {
		return errors.New("mismatched signature and payload")
	}

	return nil
}
