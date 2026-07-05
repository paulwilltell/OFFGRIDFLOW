package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

// Config holds email configuration
type Config struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromAddress  string
	FromName     string
	UseTLS       bool
}

// Client handles email sending
type Client struct {
	config     Config
	templates  *template.Template
	logger     *slog.Logger
	httpClient *http.Client
}

// NewClient creates a new email client
func NewClient(config Config, logger *slog.Logger) (*Client, error) {
	if logger == nil {
		logger = slog.Default()
	}

	// Load email templates
	templates, err := loadTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to load email templates: %w", err)
	}

	return &Client{
		config:     config,
		templates:  templates,
		logger:     logger,
		httpClient: &http.Client{Timeout: 20 * time.Second},
	}, nil
}

// Message represents an email message
type Message struct {
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	HTMLBody    string
	TextBody    string
	Attachments []Attachment
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// useSendGridAPI reports whether to send via SendGrid's HTTPS API instead of
// SMTP. Many hosts (Railway included) throttle or block outbound SMTP ports
// (465/587), which makes SMTP sends hang. SendGrid's v3 API runs over HTTPS
// (443), which is never blocked, so we prefer it whenever the configured host
// is SendGrid and we have an API key (the SMTP password doubles as the key).
func (c *Client) useSendGridAPI() bool {
	return strings.Contains(strings.ToLower(c.config.SMTPHost), "sendgrid") && c.config.SMTPPassword != ""
}

// Send sends an email message
func (c *Client) Send(ctx context.Context, msg *Message) error {
	if c.useSendGridAPI() {
		return c.sendViaSendGridAPI(ctx, msg)
	}

	// Build email message
	from := fmt.Sprintf("%s <%s>", c.config.FromName, c.config.FromAddress)
	
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))
	
	if len(msg.CC) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(msg.CC, ", ")))
	}
	
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	
	// Multipart message for HTML and text
	boundary := fmt.Sprintf("boundary_%d", time.Now().Unix())
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
	buf.WriteString("\r\n")
	
	// Text version
	if msg.TextBody != "" {
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(msg.TextBody)
		buf.WriteString("\r\n")
	}
	
	// HTML version
	if msg.HTMLBody != "" {
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(msg.HTMLBody)
		buf.WriteString("\r\n")
	}
	
	buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	// Send email
	addr := fmt.Sprintf("%s:%d", c.config.SMTPHost, c.config.SMTPPort)
	
	// Combine all recipients
	recipients := append(msg.To, msg.CC...)
	recipients = append(recipients, msg.BCC...)

	c.logger.Info("Sending email",
		"to", msg.To,
		"subject", msg.Subject,
		"smtp_host", c.config.SMTPHost)

	if c.config.UseTLS {
		return c.sendTLS(addr, recipients, buf.Bytes())
	}

	// Plain SMTP
	auth := smtp.PlainAuth("", c.config.SMTPUsername, c.config.SMTPPassword, c.config.SMTPHost)
	err := smtp.SendMail(addr, auth, c.config.FromAddress, recipients, buf.Bytes())
	if err != nil {
		c.logger.Error("Failed to send email", "error", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	c.logger.Info("Email sent successfully", "to", msg.To)
	return nil
}

func (c *Client) sendTLS(addr string, recipients []string, msg []byte) error {
	// Connect with TLS. Use a bounded dial timeout so a blocked/throttled SMTP
	// port fails fast instead of hanging the calling request indefinitely.
	tlsConfig := &tls.Config{
		ServerName: c.config.SMTPHost,
	}

	dialer := &net.Dialer{Timeout: 15 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, c.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Authenticate
	auth := smtp.PlainAuth("", c.config.SMTPUsername, c.config.SMTPPassword, c.config.SMTPHost)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Set sender
	if err := client.Mail(c.config.FromAddress); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to add recipient %s: %w", recipient, err)
		}
	}

	// Send message
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close message: %w", err)
	}

	return nil
}

// sendGridAPIURL is the SendGrid v3 mail-send endpoint. It is a package
// variable so tests can point it at a mock server.
var sendGridAPIURL = "https://api.sendgrid.com/v3/mail/send"

// sendViaSendGridAPI delivers a message through SendGrid's v3 /mail/send HTTPS
// endpoint. The API key is the configured SMTP password (SendGrid uses the same
// key for SMTP auth and API auth). This avoids outbound SMTP port blocking.
func (c *Client) sendViaSendGridAPI(ctx context.Context, msg *Message) error {
	type sgAddr struct {
		Email string `json:"email"`
		Name  string `json:"name,omitempty"`
	}
	type sgContent struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}
	type sgPersonalization struct {
		To []sgAddr `json:"to"`
	}

	to := make([]sgAddr, 0, len(msg.To))
	for _, t := range msg.To {
		to = append(to, sgAddr{Email: t})
	}

	// SendGrid requires content ordered text/plain before text/html.
	contents := make([]sgContent, 0, 2)
	if msg.TextBody != "" {
		contents = append(contents, sgContent{Type: "text/plain", Value: msg.TextBody})
	}
	if msg.HTMLBody != "" {
		contents = append(contents, sgContent{Type: "text/html", Value: msg.HTMLBody})
	}
	if len(contents) == 0 {
		contents = append(contents, sgContent{Type: "text/plain", Value: " "})
	}

	payload := map[string]interface{}{
		"personalizations": []sgPersonalization{{To: to}},
		"from":             sgAddr{Email: c.config.FromAddress, Name: c.config.FromName},
		"subject":          msg.Subject,
		"content":          contents,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal sendgrid payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sendGridAPIURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build sendgrid request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.config.SMTPPassword)
	req.Header.Set("Content-Type", "application/json")

	c.logger.Info("Sending email via SendGrid API", "to", msg.To, "subject", msg.Subject)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sendgrid api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		c.logger.Info("Email sent via SendGrid API", "to", msg.To, "status", resp.StatusCode)
		return nil
	}

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	return fmt.Errorf("sendgrid api error: status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
}

// SendPasswordReset sends a password reset email
func (c *Client) SendPasswordReset(ctx context.Context, to, name, resetToken string) error {
	resetURL := fmt.Sprintf("https://app.offgridflow.com/password/reset?token=%s", resetToken)

	data := map[string]interface{}{
		"Name":     name,
		"ResetURL": resetURL,
		"ExpiresIn": "24 hours",
	}

	var htmlBuf, textBuf bytes.Buffer
	if err := c.templates.ExecuteTemplate(&htmlBuf, "password_reset.html", data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}
	if err := c.templates.ExecuteTemplate(&textBuf, "password_reset.txt", data); err != nil {
		return fmt.Errorf("failed to execute text template: %w", err)
	}

	msg := &Message{
		To:       []string{to},
		Subject:  "Reset Your OffGridFlow Password",
		HTMLBody: htmlBuf.String(),
		TextBody: textBuf.String(),
	}

	return c.Send(ctx, msg)
}

// SendUserInvitation sends a user invitation email
func (c *Client) SendUserInvitation(ctx context.Context, to, inviterName, organizationName, inviteToken string) error {
	inviteURL := fmt.Sprintf("https://app.offgridflow.com/accept-invite?token=%s", inviteToken)

	data := map[string]interface{}{
		"InviterName":      inviterName,
		"OrganizationName": organizationName,
		"InviteURL":        inviteURL,
		"ExpiresIn":        "7 days",
	}

	var htmlBuf, textBuf bytes.Buffer
	if err := c.templates.ExecuteTemplate(&htmlBuf, "user_invitation.html", data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}
	if err := c.templates.ExecuteTemplate(&textBuf, "user_invitation.txt", data); err != nil {
		return fmt.Errorf("failed to execute text template: %w", err)
	}

	msg := &Message{
		To:       []string{to},
		Subject:  fmt.Sprintf("You've been invited to join %s on OffGridFlow", organizationName),
		HTMLBody: htmlBuf.String(),
		TextBody: textBuf.String(),
	}

	return c.Send(ctx, msg)
}

// SendWelcome sends a welcome email to new users
func (c *Client) SendWelcome(ctx context.Context, to, name string) error {
	data := map[string]interface{}{
		"Name": name,
		"DashboardURL": "https://app.offgridflow.com/dashboard",
		"DocsURL":      "https://docs.offgridflow.com",
		"SupportEmail": "support@offgridflow.com",
	}

	var htmlBuf, textBuf bytes.Buffer
	if err := c.templates.ExecuteTemplate(&htmlBuf, "welcome.html", data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}
	if err := c.templates.ExecuteTemplate(&textBuf, "welcome.txt", data); err != nil {
		return fmt.Errorf("failed to execute text template: %w", err)
	}

	msg := &Message{
		To:       []string{to},
		Subject:  "Welcome to OffGridFlow!",
		HTMLBody: htmlBuf.String(),
		TextBody: textBuf.String(),
	}

	return c.Send(ctx, msg)
}

// SendEmailVerification sends an email verification link
func (c *Client) SendEmailVerification(ctx context.Context, to, name, verifyURL string, expiresIn time.Duration) error {
	data := map[string]interface{}{
		"Name":        name,
		"VerifyURL":   verifyURL,
		"ExpiresIn":   humanizeDuration(expiresIn),
		"SupportEmail": "support@offgridflow.com",
	}

	var htmlBuf, textBuf bytes.Buffer
	if err := c.templates.ExecuteTemplate(&htmlBuf, "email_verification.html", data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}
	if err := c.templates.ExecuteTemplate(&textBuf, "email_verification.txt", data); err != nil {
		return fmt.Errorf("failed to execute text template: %w", err)
	}

	msg := &Message{
		To:       []string{to},
		Subject:  "Verify your OffGridFlow email",
		HTMLBody: htmlBuf.String(),
		TextBody: textBuf.String(),
	}

	return c.Send(ctx, msg)
}

// SendReportReady sends notification that a report is ready
func (c *Client) SendReportReady(ctx context.Context, to, name, reportType, downloadURL string) error {
	data := map[string]interface{}{
		"Name":        name,
		"ReportType":  reportType,
		"DownloadURL": downloadURL,
		"ExpiresIn":   "7 days",
	}

	var htmlBuf, textBuf bytes.Buffer
	if err := c.templates.ExecuteTemplate(&htmlBuf, "report_ready.html", data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}
	if err := c.templates.ExecuteTemplate(&textBuf, "report_ready.txt", data); err != nil {
		return fmt.Errorf("failed to execute text template: %w", err)
	}

	msg := &Message{
		To:       []string{to},
		Subject:  fmt.Sprintf("Your %s report is ready", reportType),
		HTMLBody: htmlBuf.String(),
		TextBody: textBuf.String(),
	}

	return c.Send(ctx, msg)
}

func humanizeDuration(d time.Duration) string {
	if d <= 0 {
		return "24 hours"
	}
	if d%time.Hour == 0 {
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	if d%time.Minute == 0 {
		minutes := int(d.Minutes())
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	}
	return d.String()
}

// SendTrialEnding sends notification that trial is ending soon
func (c *Client) SendTrialEnding(ctx context.Context, to, name string, daysRemaining int) error {
	data := map[string]interface{}{
		"Name":           name,
		"DaysRemaining":  daysRemaining,
		"UpgradeURL":     "https://app.offgridflow.com/settings/billing",
	}

	var htmlBuf, textBuf bytes.Buffer
	if err := c.templates.ExecuteTemplate(&htmlBuf, "trial_ending.html", data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}
	if err := c.templates.ExecuteTemplate(&textBuf, "trial_ending.txt", data); err != nil {
		return fmt.Errorf("failed to execute text template: %w", err)
	}

	msg := &Message{
		To:       []string{to},
		Subject:  fmt.Sprintf("Your OffGridFlow trial ends in %d days", daysRemaining),
		HTMLBody: htmlBuf.String(),
		TextBody: textBuf.String(),
	}

	return c.Send(ctx, msg)
}

// SendPaymentFailed sends notification about payment failure
func (c *Client) SendPaymentFailed(ctx context.Context, to, name string, amountCents int64) error {
	data := map[string]interface{}{
		"Name":            name,
		"Amount":          fmt.Sprintf("$%.2f", float64(amountCents)/100),
		"UpdatePaymentURL": "https://app.offgridflow.com/settings/billing",
	}

	var htmlBuf, textBuf bytes.Buffer
	if err := c.templates.ExecuteTemplate(&htmlBuf, "payment_failed.html", data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}
	if err := c.templates.ExecuteTemplate(&textBuf, "payment_failed.txt", data); err != nil {
		return fmt.Errorf("failed to execute text template: %w", err)
	}

	msg := &Message{
		To:       []string{to},
		Subject:  "Payment Failed - Action Required",
		HTMLBody: htmlBuf.String(),
		TextBody: textBuf.String(),
	}

	return c.Send(ctx, msg)
}

func loadTemplates() (*template.Template, error) {
	// In production, load from files or embed
	// For now, create inline templates
	tmpl := template.New("email")

	// Password reset HTML template
	template.Must(tmpl.New("password_reset.html").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Reset Your Password</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2>Reset Your Password</h2>
        <p>Hi {{.Name}},</p>
        <p>We received a request to reset your password for your OffGridFlow account.</p>
        <p>Click the button below to reset your password:</p>
        <p style="margin: 30px 0;">
            <a href="{{.ResetURL}}" style="background-color: #4CAF50; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Reset Password</a>
        </p>
        <p>This link will expire in {{.ExpiresIn}}.</p>
        <p>If you didn't request this password reset, you can safely ignore this email.</p>
        <p>Best regards,<br>The OffGridFlow Team</p>
    </div>
</body>
</html>
	`))

	// Password reset text template
	template.Must(tmpl.New("password_reset.txt").Parse(`
Reset Your Password

Hi {{.Name}},

We received a request to reset your password for your OffGridFlow account.

Click the link below to reset your password:
{{.ResetURL}}

This link will expire in {{.ExpiresIn}}.

If you didn't request this password reset, you can safely ignore this email.

Best regards,
The OffGridFlow Team
	`))

	// Email verification templates
	template.Must(tmpl.New("email_verification.html").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verify your email</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2>Verify your email</h2>
        <p>Hi {{.Name}},</p>
        <p>Thanks for creating your OffGridFlow account. Please verify your email to finish setting up your account.</p>
        <p style="margin: 30px 0;">
            <a href="{{.VerifyURL}}" style="background-color: #16a34a; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Verify Email</a>
        </p>
        <p>This link will expire in {{.ExpiresIn}}.</p>
        <p>If you did not create this account, you can ignore this email or contact {{.SupportEmail}}.</p>
        <p>Best regards,<br>The OffGridFlow Team</p>
    </div>
</body>
</html>
	`))

	template.Must(tmpl.New("email_verification.txt").Parse(`
Verify your email

Hi {{.Name}},

Thanks for creating your OffGridFlow account. Please verify your email to finish setting up your account:
{{.VerifyURL}}

This link will expire in {{.ExpiresIn}}.

If you did not create this account, you can ignore this email or contact {{.SupportEmail}}.

Best regards,
The OffGridFlow Team
	`))

	// User invitation HTML template
	template.Must(tmpl.New("user_invitation.html").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>You're Invited!</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2>You're Invited to Join {{.OrganizationName}}</h2>
        <p>{{.InviterName}} has invited you to join {{.OrganizationName}} on OffGridFlow.</p>
        <p>Click the button below to accept the invitation:</p>
        <p style="margin: 30px 0;">
            <a href="{{.InviteURL}}" style="background-color: #4CAF50; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Accept Invitation</a>
        </p>
        <p>This invitation will expire in {{.ExpiresIn}}.</p>
        <p>Best regards,<br>The OffGridFlow Team</p>
    </div>
</body>
</html>
	`))

	template.Must(tmpl.New("user_invitation.txt").Parse(`
You're Invited to Join {{.OrganizationName}}

{{.InviterName}} has invited you to join {{.OrganizationName}} on OffGridFlow.

Click the link below to accept the invitation:
{{.InviteURL}}

This invitation will expire in {{.ExpiresIn}}.

Best regards,
The OffGridFlow Team
	`))

	// Welcome email templates
	template.Must(tmpl.New("welcome.html").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Welcome to OffGridFlow</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2>Welcome to OffGridFlow!</h2>
        <p>Hi {{.Name}},</p>
        <p>Thank you for signing up for OffGridFlow! We're excited to help you track and reduce your carbon emissions.</p>
        <h3>Get Started</h3>
        <ul>
            <li><a href="{{.DashboardURL}}">Visit your dashboard</a></li>
            <li><a href="{{.DocsURL}}">Read the documentation</a></li>
            <li>Connect your data sources</li>
            <li>Generate your first compliance report</li>
        </ul>
        <p>If you have any questions, don't hesitate to reach out to our support team at {{.SupportEmail}}.</p>
        <p>Best regards,<br>The OffGridFlow Team</p>
    </div>
</body>
</html>
	`))

	template.Must(tmpl.New("welcome.txt").Parse(`
Welcome to OffGridFlow!

Hi {{.Name}},

Thank you for signing up for OffGridFlow! We're excited to help you track and reduce your carbon emissions.

Get Started:
- Visit your dashboard: {{.DashboardURL}}
- Read the documentation: {{.DocsURL}}
- Connect your data sources
- Generate your first compliance report

If you have any questions, don't hesitate to reach out to our support team at {{.SupportEmail}}.

Best regards,
The OffGridFlow Team
	`))

	// Report ready templates
	template.Must(tmpl.New("report_ready.html").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Your Report is Ready</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2>Your {{.ReportType}} Report is Ready</h2>
        <p>Hi {{.Name}},</p>
        <p>Your {{.ReportType}} report has been generated and is ready for download.</p>
        <p style="margin: 30px 0;">
            <a href="{{.DownloadURL}}" style="background-color: #4CAF50; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Download Report</a>
        </p>
        <p>This download link will be available for {{.ExpiresIn}}.</p>
        <p>Best regards,<br>The OffGridFlow Team</p>
    </div>
</body>
</html>
	`))

	template.Must(tmpl.New("report_ready.txt").Parse(`
Your {{.ReportType}} Report is Ready

Hi {{.Name}},

Your {{.ReportType}} report has been generated and is ready for download.

Download your report here:
{{.DownloadURL}}

This download link will be available for {{.ExpiresIn}}.

Best regards,
The OffGridFlow Team
	`))

	// Trial ending templates
	template.Must(tmpl.New("trial_ending.html").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Your Trial is Ending Soon</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2>Your Trial is Ending Soon</h2>
        <p>Hi {{.Name}},</p>
        <p>Your OffGridFlow trial will end in {{.DaysRemaining}} days.</p>
        <p>To continue using OffGridFlow and keep access to all your data and reports, please upgrade to a paid plan.</p>
        <p style="margin: 30px 0;">
            <a href="{{.UpgradeURL}}" style="background-color: #4CAF50; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Upgrade Now</a>
        </p>
        <p>Best regards,<br>The OffGridFlow Team</p>
    </div>
</body>
</html>
	`))

	template.Must(tmpl.New("trial_ending.txt").Parse(`
Your Trial is Ending Soon

Hi {{.Name}},

Your OffGridFlow trial will end in {{.DaysRemaining}} days.

To continue using OffGridFlow and keep access to all your data and reports, please upgrade to a paid plan.

Upgrade now: {{.UpgradeURL}}

Best regards,
The OffGridFlow Team
	`))

	// Payment failed templates
	template.Must(tmpl.New("payment_failed.html").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Payment Failed</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2>Payment Failed - Action Required</h2>
        <p>Hi {{.Name}},</p>
        <p>We were unable to process your payment of {{.Amount}} for your OffGridFlow subscription.</p>
        <p>Please update your payment method to avoid service interruption.</p>
        <p style="margin: 30px 0;">
            <a href="{{.UpdatePaymentURL}}" style="background-color: #f44336; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Update Payment Method</a>
        </p>
        <p>Best regards,<br>The OffGridFlow Team</p>
    </div>
</body>
</html>
	`))

	template.Must(tmpl.New("payment_failed.txt").Parse(`
Payment Failed - Action Required

Hi {{.Name}},

We were unable to process your payment of {{.Amount}} for your OffGridFlow subscription.

Please update your payment method to avoid service interruption.

Update payment method: {{.UpdatePaymentURL}}

Best regards,
The OffGridFlow Team
	`))

	return tmpl, nil
}
