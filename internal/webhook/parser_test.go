package webhook_test

import (
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/webhook"
	"os"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	payload, err := os.ReadFile("testdata/webhook.json")
	if err != nil {
		t.Fatal(err)
	}
	event, err := webhook.Parse(payload)
	if err != nil {
		t.Fatal(err)
	}
	if event.Id != 1001 || event.Status != webhook.Active {
		t.Fatal("unexpected synthetic event")
	}
}

func TestMsisdnExtraction(t *testing.T) {
	payload, err := os.ReadFile("testdata/webhook.json")
	if err != nil {
		t.Fatal(err)
	}
	event, err := webhook.Parse(payload)
	if err != nil {
		t.Fatal(err)
	}
	msisdn, err := event.Msisdn()
	if err != nil {
		t.Fatal(err)
	}
	if msisdn != "+12025550123" {
		t.Fatal("unexpected synthetic phone number")
	}
}

func TestEmailExtraction(t *testing.T) {
	payload, err := os.ReadFile("testdata/webhook.json")
	if err != nil {
		t.Fatal(err)
	}
	event, err := webhook.Parse(payload)
	if err != nil {
		t.Fatal(err)
	}
	email := event.Email
	if email != "subscriber@example.invalid" {
		t.Fatal("unexpected synthetic email")
	}
}

func TestCancellationExpirationTimeFromField(t *testing.T) {
	e := &webhook.Event{
		CancellationScheduledFor: "2026-04-20T22:03:26+00:00",
	}
	got, err := e.CancellationExpirationTime()
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 4, 20, 22, 3, 26, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestCancellationExpirationTimeFromCanceledJSON(t *testing.T) {
	payload, err := os.ReadFile("testdata/canceled.json")
	if err != nil {
		t.Fatal(err)
	}
	event, err := webhook.Parse(payload)
	if err != nil {
		t.Fatal(err)
	}
	got, err := event.CancellationExpirationTime()
	if err != nil {
		t.Fatal(err)
	}
	// cancellation_scheduled_for = "2026-04-20T22:03:26+00:00"
	want := time.Date(2026, 4, 20, 22, 3, 26, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestParseNullRejected(t *testing.T) {
	if _, err := webhook.Parse([]byte("null")); err == nil {
		t.Fatal("null event must be rejected")
	}
}

func TestCancellationExpirationTimeEmptyField(t *testing.T) {
	e := &webhook.Event{}
	before := time.Now().UTC()
	got, err := e.CancellationExpirationTime()
	after := time.Now().UTC()
	if err != nil {
		t.Fatalf("expected no error when cancellation_scheduled_for is empty, got %v", err)
	}
	if got.Before(before) || got.After(after) {
		t.Fatalf("expected fallback time between %s and %s, got %s", before, after, got)
	}
}
