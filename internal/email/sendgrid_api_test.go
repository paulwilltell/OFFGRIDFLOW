package email

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// When the host is SendGrid, Send must POST to the v3 API over HTTPS with the
// API key as a Bearer token and a correctly-shaped payload — never touch SMTP.
func TestSendGridAPISend(t *testing.T) {
	var gotAuth, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusAccepted) // SendGrid returns 202
	}))
	defer srv.Close()

	orig := sendGridAPIURL
	sendGridAPIURL = srv.URL
	defer func() { sendGridAPIURL = orig }()

	c, err := NewClient(Config{
		SMTPHost:     "smtp.sendgrid.net",
		SMTPUsername: "apikey",
		SMTPPassword: "SG.test-key-123",
		FromAddress:  "noreply@off-grid-flow.com",
		FromName:     "OffGridFlow",
		UseTLS:       true,
	}, nil)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if !c.useSendGridAPI() {
		t.Fatal("expected useSendGridAPI() true for smtp.sendgrid.net host")
	}

	err = c.Send(context.Background(), &Message{
		To:       []string{"user@example.com"},
		Subject:  "Verify your OffGridFlow email",
		TextBody: "verify here",
		HTMLBody: "<a>verify</a>",
	})
	if err != nil {
		t.Fatalf("send via api: %v", err)
	}

	if gotAuth != "Bearer SG.test-key-123" {
		t.Errorf("expected Bearer auth with API key, got %q", gotAuth)
	}

	// Payload must carry recipient, from, subject, and both content parts.
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(gotBody), &payload); err != nil {
		t.Fatalf("payload not valid JSON: %v", err)
	}
	if !strings.Contains(gotBody, "user@example.com") {
		t.Errorf("payload missing recipient: %s", gotBody)
	}
	if !strings.Contains(gotBody, "text/plain") || !strings.Contains(gotBody, "text/html") {
		t.Errorf("payload missing content parts: %s", gotBody)
	}
	if payload["subject"] != "Verify your OffGridFlow email" {
		t.Errorf("payload subject wrong: %v", payload["subject"])
	}
}

// A non-2xx from SendGrid must surface as an error (so callers can log it).
func TestSendGridAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":[{"message":"invalid key"}]}`))
	}))
	defer srv.Close()

	orig := sendGridAPIURL
	sendGridAPIURL = srv.URL
	defer func() { sendGridAPIURL = orig }()

	c, _ := NewClient(Config{
		SMTPHost:     "smtp.sendgrid.net",
		SMTPPassword: "SG.bad",
		FromAddress:  "noreply@off-grid-flow.com",
	}, nil)

	err := c.Send(context.Background(), &Message{To: []string{"u@example.com"}, Subject: "x", TextBody: "y"})
	if err == nil {
		t.Fatal("expected error on 401 from SendGrid")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected status in error, got %v", err)
	}
}
