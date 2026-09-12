package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestVerifierRejectsEmptyKey(t *testing.T) {
	if _, err := NewVerifier(""); err == nil {
		t.Fatal("empty signing key must be rejected")
	}
}

func TestVerifierAcceptsExactPayloadOnly(t *testing.T) {
	key := "synthetic-signing-key-not-for-use"
	payload := []byte(`{"id":1001}`)
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(payload)
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	v, err := NewVerifier(key)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Verify(signature, payload); err != nil {
		t.Fatal("valid synthetic signature rejected")
	}
	if err := v.Verify(signature, []byte(`{"id":1002}`)); err == nil {
		t.Fatal("modified payload accepted")
	}
	if err := v.Verify("", payload); err == nil {
		t.Fatal("empty signature accepted")
	}
}
