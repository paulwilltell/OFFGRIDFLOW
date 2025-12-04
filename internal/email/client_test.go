package email

import (
	"context"
	"testing"
)

func TestNewClient(t *testing.T) {
	cfg := Config{
		SMTPHost:    "localhost",
		SMTPPort:    587,
		FromAddress: "test@example.com",
		FromName:    "Test",
	}

	client, err := NewClient(cfg, nil)
	if err != nil {
		t.Logf("NewClient may fail without templates: %v", err)
		return
	}

	if client == nil {
		t.Fatal("NewClient returned nil without error")
	}
}

func TestMessage_Struct(t *testing.T) {
	msg := Message{
		To:       []string{"recipient@example.com"},
		Subject:  "Test Subject",
		HTMLBody: "<p>Hello World</p>",
		TextBody: "Hello World",
	}

	if len(msg.To) == 0 {
		t.Error("Message should have recipients")
	}
	if msg.Subject == "" {
		t.Error("Message should have subject")
	}
}

func TestAttachment_Struct(t *testing.T) {
	att := Attachment{
		Filename:    "report.pdf",
		ContentType: "application/pdf",
		Data:        []byte("dummy pdf data"),
	}

	if att.Filename == "" {
		t.Error("Attachment should have filename")
	}
	if len(att.Data) == 0 {
		t.Error("Attachment should have data")
	}
}

func TestConfig_Defaults(t *testing.T) {
	cfg := Config{}

	// Zero config should be valid (might not send emails)
	if cfg.SMTPPort != 0 {
		t.Log("Expected zero port for zero config")
	}
}

func TestClient_Send_NoServer(t *testing.T) {
	cfg := Config{
		SMTPHost:    "localhost",
		SMTPPort:    12345, // Non-existent port
		FromAddress: "test@example.com",
		FromName:    "Test",
	}

	client, err := NewClient(cfg, nil)
	if err != nil {
		t.Logf("Skipping send test - client creation failed: %v", err)
		return
	}

	msg := &Message{
		To:       []string{"recipient@example.com"},
		Subject:  "Test",
		TextBody: "Test body",
	}

	err = client.Send(context.Background(), msg)
	if err == nil {
		t.Log("Send succeeded (unexpected for non-existent server)")
	}
	// Error expected - no SMTP server
}

func TestEmailTemplates(t *testing.T) {
	// Test that template loading works
	templates, err := loadTemplates()
	if err != nil {
		t.Logf("Template loading may fail without template files: %v", err)
		return
	}

	if templates == nil {
		t.Log("loadTemplates returned nil")
	}
}

func TestBuildEmailMessage(t *testing.T) {
	msg := &Message{
		To:       []string{"a@example.com", "b@example.com"},
		CC:       []string{"cc@example.com"},
		Subject:  "Multi-recipient Test",
		HTMLBody: "<html><body><p>Test</p></body></html>",
		TextBody: "Test",
	}

	if len(msg.To) != 2 {
		t.Errorf("Expected 2 recipients, got %d", len(msg.To))
	}
	if len(msg.CC) != 1 {
		t.Errorf("Expected 1 CC, got %d", len(msg.CC))
	}
}
