package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/example/offgridflow/internal/email"
	"github.com/stripe/stripe-go/v82"
)

// WebhookHandler handles Stripe webhook events
type WebhookHandler struct {
	stripeClient *StripeClient
	service      *Service
	logger       *slog.Logger
	emailClient  *email.Client
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(stripeClient *StripeClient, service *Service, logger *slog.Logger) *WebhookHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &WebhookHandler{
		stripeClient: stripeClient,
		service:      service,
		logger:       logger,
	}
}

// WithEmailClient attaches an email client for customer notifications.
func (h *WebhookHandler) WithEmailClient(client *email.Client) *WebhookHandler {
	h.emailClient = client
	return h
}

// HandleWebhook processes incoming Stripe webhooks
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse and verify webhook event
	event, err := h.stripeClient.ParseWebhook(r)
	if err != nil {
		h.logger.Error("Failed to parse webhook", "error", err)
		http.Error(w, "Invalid signature", http.StatusBadRequest)
		return
	}

	h.logger.Info("Received webhook", "type", event.Type, "id", event.ID)

	// Route to appropriate handler
	var handlerErr error
	switch event.Type {
	case "customer.created":
		handlerErr = h.handleCustomerCreated(ctx, event)
	case "customer.updated":
		handlerErr = h.handleCustomerUpdated(ctx, event)
	case "customer.deleted":
		handlerErr = h.handleCustomerDeleted(ctx, event)
	case "customer.subscription.created":
		handlerErr = h.handleSubscriptionCreated(ctx, event)
	case "customer.subscription.updated":
		handlerErr = h.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		handlerErr = h.handleSubscriptionDeleted(ctx, event)
	case "customer.subscription.trial_will_end":
		handlerErr = h.handleTrialWillEnd(ctx, event)
	case "invoice.created":
		handlerErr = h.handleInvoiceCreated(ctx, event)
	case "invoice.paid":
		handlerErr = h.handleInvoicePaid(ctx, event)
	case "invoice.payment_failed":
		handlerErr = h.handleInvoicePaymentFailed(ctx, event)
	case "payment_intent.succeeded":
		handlerErr = h.handlePaymentSucceeded(ctx, event)
	case "payment_intent.payment_failed":
		handlerErr = h.handlePaymentFailed(ctx, event)
	case "checkout.session.completed":
		handlerErr = h.handleCheckoutCompleted(ctx, event)
	default:
		h.logger.Info("Unhandled webhook type", "type", event.Type)
	}

	if handlerErr != nil {
		h.logger.Error("Webhook handler error", "type", event.Type, "error", handlerErr)
		http.Error(w, "Handler error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *WebhookHandler) handleCustomerCreated(ctx context.Context, event *stripe.Event) error {
	var customer stripe.Customer
	if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
		return fmt.Errorf("failed to parse customer: %w", err)
	}

	h.logger.Info("Customer created", "customer_id", customer.ID, "email", customer.Email)

	tenantID, ok := customer.Metadata["tenant_id"]
	if !ok {
		return fmt.Errorf("customer missing tenant_id metadata")
	}

	// Update tenant with Stripe customer ID
	if err := h.service.UpdateTenantStripeCustomer(ctx, tenantID, customer.ID); err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	return nil
}

func (h *WebhookHandler) handleCustomerUpdated(ctx context.Context, event *stripe.Event) error {
	var customer stripe.Customer
	if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
		return fmt.Errorf("failed to parse customer: %w", err)
	}

	h.logger.Info("Customer updated", "customer_id", customer.ID)
	return nil
}

func (h *WebhookHandler) handleCustomerDeleted(ctx context.Context, event *stripe.Event) error {
	var customer stripe.Customer
	if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
		return fmt.Errorf("failed to parse customer: %w", err)
	}

	h.logger.Info("Customer deleted", "customer_id", customer.ID)

	tenantID, ok := customer.Metadata["tenant_id"]
	if !ok {
		return nil
	}

	// Clear Stripe customer ID from tenant
	if err := h.service.UpdateTenantStripeCustomer(ctx, tenantID, ""); err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	return nil
}

func (h *WebhookHandler) handleSubscriptionCreated(ctx context.Context, event *stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	// Validate required fields
	if sub.Customer == nil {
		return fmt.Errorf("subscription missing customer")
	}
	if sub.Items == nil || len(sub.Items.Data) == 0 || sub.Items.Data[0].Price == nil {
		return fmt.Errorf("subscription missing price information")
	}

	h.logger.Info("Subscription created",
		"subscription_id", sub.ID,
		"customer_id", sub.Customer.ID,
		"status", sub.Status)

	tenantID, ok := sub.Metadata["tenant_id"]
	if !ok {
		return fmt.Errorf("subscription missing tenant_id metadata")
	}

	// Determine plan tier from price ID
	planTier := h.getPlanFromPriceID(sub.Items.Data[0].Price.ID)

	// Update tenant subscription
	if err := h.service.UpdateTenantSubscription(ctx, tenantID, sub.ID, string(planTier), string(sub.Status)); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

func (h *WebhookHandler) handleSubscriptionUpdated(ctx context.Context, event *stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	// Validate required fields
	if sub.Items == nil || len(sub.Items.Data) == 0 || sub.Items.Data[0].Price == nil {
		return fmt.Errorf("subscription missing price information")
	}

	h.logger.Info("Subscription updated",
		"subscription_id", sub.ID,
		"status", sub.Status,
		"cancel_at_period_end", sub.CancelAtPeriodEnd)

	tenantID, ok := sub.Metadata["tenant_id"]
	if !ok {
		return fmt.Errorf("subscription missing tenant_id metadata")
	}

	planTier := h.getPlanFromPriceID(sub.Items.Data[0].Price.ID)

	// Extract period end from subscription item
	var periodEnd *time.Time
	if sub.Items.Data[0].CurrentPeriodEnd != 0 {
		t := time.Unix(sub.Items.Data[0].CurrentPeriodEnd, 0)
		periodEnd = &t
	}

	if err := h.service.UpdateTenantSubscriptionWithPeriod(ctx, tenantID, sub.ID, string(planTier), string(sub.Status), periodEnd); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

func (h *WebhookHandler) handleSubscriptionDeleted(ctx context.Context, event *stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	h.logger.Info("Subscription deleted", "subscription_id", sub.ID)

	tenantID, ok := sub.Metadata["tenant_id"]
	if !ok {
		return fmt.Errorf("subscription missing tenant_id metadata")
	}

	// Downgrade to free plan
	if err := h.service.UpdateTenantSubscription(ctx, tenantID, "", "free", "canceled"); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

func (h *WebhookHandler) handleTrialWillEnd(ctx context.Context, event *stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	h.logger.Info("Trial ending soon",
		"subscription_id", sub.ID,
		"trial_end", sub.TrialEnd)

	// Send customer notification if email is available
	if sub.Customer != nil && sub.Customer.Email != "" {
		subject := "Your OffGridFlow trial is ending soon"
		body := fmt.Sprintf("Your trial will end on %v. Please add payment details to avoid interruption.", sub.TrialEnd)
		h.sendEmail(ctx, sub.Customer.Email, subject, body)
	}

	return nil
}

func (h *WebhookHandler) handleInvoiceCreated(ctx context.Context, event *stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("failed to parse invoice: %w", err)
	}

	h.logger.Info("Invoice created",
		"invoice_id", invoice.ID,
		"customer_id", invoice.Customer.ID,
		"amount", invoice.AmountDue)

	return nil
}

func (h *WebhookHandler) handleInvoicePaid(ctx context.Context, event *stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("failed to parse invoice: %w", err)
	}

	h.logger.Info("Invoice paid",
		"invoice_id", invoice.ID,
		"customer_id", invoice.Customer.ID,
		"amount_paid", invoice.AmountPaid)

	if invoice.CustomerEmail != "" {
		subject := "Payment receipt"
		body := fmt.Sprintf("We received your payment of %.2f %s. Invoice: %s", float64(invoice.AmountPaid)/100, invoice.Currency, invoice.ID)
		h.sendEmail(ctx, invoice.CustomerEmail, subject, body)
	}

	// Ensure subscription state is synchronized
	if err := h.service.HandleWebhookEvent(ctx, event); err != nil {
		return err
	}

	return nil
}

func (h *WebhookHandler) handleInvoicePaymentFailed(ctx context.Context, event *stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("failed to parse invoice: %w", err)
	}

	h.logger.Error("Invoice payment failed",
		"invoice_id", invoice.ID,
		"customer_id", invoice.Customer.ID,
		"amount_due", invoice.AmountDue)

	if invoice.CustomerEmail != "" {
		subject := "Payment failed"
		body := fmt.Sprintf("We could not process your payment for invoice %s. Please update your payment method.", invoice.ID)
		h.sendEmail(ctx, invoice.CustomerEmail, subject, body)
	}

	// Mark subscription as past due in our store
	if err := h.service.HandleWebhookEvent(ctx, event); err != nil {
		return err
	}

	return nil
}

func (h *WebhookHandler) handlePaymentSucceeded(ctx context.Context, event *stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return fmt.Errorf("failed to parse payment intent: %w", err)
	}

	h.logger.Info("Payment succeeded",
		"payment_intent_id", paymentIntent.ID,
		"amount", paymentIntent.Amount)

	return nil
}

func (h *WebhookHandler) handlePaymentFailed(ctx context.Context, event *stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return fmt.Errorf("failed to parse payment intent: %w", err)
	}

	h.logger.Error("Payment failed",
		"payment_intent_id", paymentIntent.ID,
		"last_payment_error", paymentIntent.LastPaymentError)

	return nil
}

func (h *WebhookHandler) handleCheckoutCompleted(ctx context.Context, event *stripe.Event) error {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return fmt.Errorf("failed to parse checkout session: %w", err)
	}

	h.logger.Info("Checkout completed",
		"session_id", session.ID,
		"customer_id", session.Customer.ID,
		"subscription_id", session.Subscription.ID)

	// Subscription is already created, just log success
	return nil
}

func (h *WebhookHandler) getPlanFromPriceID(priceID string) PlanTier {
	switch priceID {
	case h.stripeClient.priceFree:
		return PlanFree
	case h.stripeClient.priceBasic:
		return PlanBasic
	case h.stripeClient.pricePro:
		return PlanPro
	case h.stripeClient.priceEnterprise:
		return PlanEnterprise
	default:
		return PlanFree
	}
}

// sendEmail dispatches an email if a client is configured; otherwise logs.
func (h *WebhookHandler) sendEmail(ctx context.Context, to, subject, body string) {
	if h.emailClient == nil || to == "" {
		h.logger.Debug("email client not configured; skipping notification", "to", to, "subject", subject)
		return
	}

	msg := &email.Message{
		To:       []string{to},
		Subject:  subject,
		TextBody: body,
		HTMLBody: fmt.Sprintf("<p>%s</p>", body),
	}
	if err := h.emailClient.Send(ctx, msg); err != nil {
		h.logger.Warn("failed to send notification email", "to", to, "error", err)
	}
}
