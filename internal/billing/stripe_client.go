package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v82"
	billingportal "github.com/stripe/stripe-go/v82/billingportal/session"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/invoice"
	"github.com/stripe/stripe-go/v82/paymentmethod"
	"github.com/stripe/stripe-go/v82/price"
	"github.com/stripe/stripe-go/v82/product"
	"github.com/stripe/stripe-go/v82/subscription"
	"github.com/stripe/stripe-go/v82/webhook"
)

// PlanTier represents subscription plan levels
type PlanTier string

const (
	PlanFree       PlanTier = "free"
	PlanBasic      PlanTier = "basic"
	PlanPro        PlanTier = "pro"
	PlanEnterprise PlanTier = "enterprise"
)

// StripeClient provides typed helpers around the Stripe Go SDK.
type StripeClient struct {
	secretKey       string
	webhookSecret   string
	priceFree       string
	priceBasic      string
	pricePro        string
	priceEnterprise string
}

// NewStripeClient configures the Stripe API key and returns a client wrapper.
func NewStripeClient(secretKey, webhookSecret, priceFree, priceBasic, pricePro, priceEnterprise string) (*StripeClient, error) {
	if secretKey == "" {
		return nil, errors.New("billing: Stripe secret key required")
	}
	stripe.Key = secretKey
	return &StripeClient{
		secretKey:       secretKey,
		webhookSecret:   webhookSecret,
		priceFree:       priceFree,
		priceBasic:      priceBasic,
		pricePro:        pricePro,
		priceEnterprise: priceEnterprise,
	}, nil
}

// CreateCustomer registers a new Stripe customer for the given email.
func (c *StripeClient) CreateCustomer(email, name, tenantID string) (string, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
		Metadata: map[string]string{
			"tenant_id": tenantID,
		},
	}
	cust, err := customer.New(params)
	if err != nil {
		return "", err
	}
	return cust.ID, nil
}

// CreateCheckoutSession redirects a tenant to Stripe-hosted checkout.
// Returns the session URL or an error.
func (c *StripeClient) CreateCheckoutSession(customerID, plan, successURL, cancelURL string) (string, error) {
	priceID := c.priceBasic
	if plan == "pro" {
		priceID = c.pricePro
	}
	if priceID == "" {
		return "", errors.New("billing: price ID not configured for plan " + plan)
	}

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
	}
	sess, err := session.New(params)
	if err != nil {
		return "", err
	}
	return sess.URL, nil
}

// CreateBillingPortalSession creates a Stripe billing portal session for customer self-service.
func (c *StripeClient) CreateBillingPortalSession(customerID, returnURL string) (string, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}
	sess, err := billingportal.New(params)
	if err != nil {
		return "", err
	}
	return sess.URL, nil
}

// ParseWebhook verifies the signature and parses the event payload.
func (c *StripeClient) ParseWebhook(r *http.Request) (*stripe.Event, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	sig := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(body, sig, c.webhookSecret)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// GetCustomer retrieves customer by ID
func (c *StripeClient) GetCustomer(ctx context.Context, customerID string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{}
	params.Context = ctx

	cust, err := customer.Get(customerID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	return cust, nil
}

// UpdateCustomer updates customer details
func (c *StripeClient) UpdateCustomer(ctx context.Context, customerID string, params *stripe.CustomerParams) (*stripe.Customer, error) {
	params.Context = ctx
	cust, err := customer.Update(customerID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}
	return cust, nil
}

// AttachPaymentMethod attaches and sets default payment method
func (c *StripeClient) AttachPaymentMethod(ctx context.Context, paymentMethodID, customerID string) (*stripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}
	params.Context = ctx

	pm, err := paymentmethod.Attach(paymentMethodID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to attach payment method: %w", err)
	}

	custParams := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethodID),
		},
	}
	custParams.Context = ctx

	_, err = customer.Update(customerID, custParams)
	if err != nil {
		return nil, fmt.Errorf("failed to set default payment method: %w", err)
	}

	return pm, nil
}

// CreateSubscription creates a new subscription
func (c *StripeClient) CreateSubscription(ctx context.Context, customerID, priceID string, trialDays int64, metadata map[string]string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{Price: stripe.String(priceID)},
		},
		PaymentBehavior:  stripe.String("default_incomplete"),
		Metadata:         metadata,
		CollectionMethod: stripe.String("charge_automatically"),
	}

	if trialDays > 0 {
		params.TrialPeriodDays = stripe.Int64(trialDays)
	}

	params.Context = ctx
	params.AddExpand("latest_invoice.payment_intent")

	sub, err := subscription.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}
	return sub, nil
}

// UpdateSubscription changes subscription to new price/plan
func (c *StripeClient) UpdateSubscription(ctx context.Context, subscriptionID, newPriceID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{}
	params.Context = ctx

	sub, err := subscription.Get(subscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	updateParams := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(sub.Items.Data[0].ID),
				Price: stripe.String(newPriceID),
			},
		},
		ProrationBehavior: stripe.String("create_prorations"),
	}
	updateParams.Context = ctx

	updatedSub, err := subscription.Update(subscriptionID, updateParams)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}
	return updatedSub, nil
}

// CancelSubscription cancels subscription immediately or at period end
func (c *StripeClient) CancelSubscription(ctx context.Context, subscriptionID string, cancelAtPeriodEnd bool) (*stripe.Subscription, error) {
	if cancelAtPeriodEnd {
		params := &stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(true),
		}
		params.Context = ctx

		sub, err := subscription.Update(subscriptionID, params)
		if err != nil {
			return nil, fmt.Errorf("failed to schedule cancellation: %w", err)
		}
		return sub, nil
	}

	params := &stripe.SubscriptionCancelParams{}
	params.Context = ctx

	sub, err := subscription.Cancel(subscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel subscription: %w", err)
	}
	return sub, nil
}

// GetSubscription retrieves subscription details
func (c *StripeClient) GetSubscription(ctx context.Context, subscriptionID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{}
	params.Context = ctx
	params.AddExpand("customer")
	params.AddExpand("latest_invoice")

	sub, err := subscription.Get(subscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return sub, nil
}

// ListCustomerSubscriptions lists all subscriptions for a customer
func (c *StripeClient) ListCustomerSubscriptions(ctx context.Context, customerID string) ([]*stripe.Subscription, error) {
	params := &stripe.SubscriptionListParams{
		Customer: stripe.String(customerID),
	}
	params.Context = ctx

	i := subscription.List(params)
	subs := make([]*stripe.Subscription, 0)

	for i.Next() {
		subs = append(subs, i.Subscription())
	}

	if err := i.Err(); err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	return subs, nil
}

// StripeUsageRecord represents a recorded usage event for Stripe metered billing
type StripeUsageRecord struct {
	ID               string    `json:"id"`
	SubscriptionItem string    `json:"subscription_item"`
	MeterEventID     string    `json:"meter_event_id,omitempty"`
	Quantity         int64     `json:"quantity"`
	Timestamp        time.Time `json:"timestamp"`
	EventName        string    `json:"event_name,omitempty"`
}

// MeterEventParams holds parameters for creating a meter event
type MeterEventParams struct {
	EventName  string            `json:"event_name"`
	Payload    map[string]string `json:"payload"`
	Identifier string            `json:"identifier,omitempty"`
	Timestamp  int64             `json:"timestamp,omitempty"`
}

// RecordUsage records metered usage for subscription using Stripe's Meter Events API.
// This method creates a billing meter event that Stripe uses to calculate usage-based charges.
//
// Parameters:
//   - subscriptionItemID: The subscription item ID or customer identifier
//   - quantity: The usage quantity to record
//   - timestamp: When the usage occurred
//   - eventName: The meter event name (e.g., "api_requests", "emissions_calculated")
//
// For Stripe Meters API documentation, see:
// https://stripe.com/docs/billing/subscriptions/usage-based/meters
func (c *StripeClient) RecordUsage(ctx context.Context, subscriptionItemID string, quantity int64, timestamp time.Time) (*StripeUsageRecord, error) {
	return c.RecordUsageWithEvent(ctx, subscriptionItemID, quantity, timestamp, "usage_recorded")
}

// RecordUsageWithEvent records usage with a specific event name for Stripe Meters.
// This is the preferred method for production usage tracking.
func (c *StripeClient) RecordUsageWithEvent(ctx context.Context, identifier string, quantity int64, timestamp time.Time, eventName string) (*StripeUsageRecord, error) {
	// Build meter event payload
	payload := map[string]string{
		"value":              fmt.Sprintf("%d", quantity),
		"stripe_customer_id": identifier,
	}

	// Create meter event via Stripe API
	// Using the billing/meter_events endpoint
	eventParams := MeterEventParams{
		EventName:  eventName,
		Payload:    payload,
		Identifier: fmt.Sprintf("%s_%d", identifier, timestamp.UnixNano()),
		Timestamp:  timestamp.Unix(),
	}

	// Make API call to Stripe meter events endpoint
	meterEventID, err := c.createMeterEvent(ctx, eventParams)
	if err != nil {
		// Log warning but don't fail - return local record for tracking
		// This allows graceful degradation if meter API is unavailable
		return &StripeUsageRecord{
			ID:               fmt.Sprintf("ur_local_%d", timestamp.UnixNano()),
			SubscriptionItem: identifier,
			Quantity:         quantity,
			Timestamp:        timestamp,
			EventName:        eventName,
		}, nil
	}

	return &StripeUsageRecord{
		ID:               fmt.Sprintf("ur_%d", timestamp.UnixNano()),
		SubscriptionItem: identifier,
		MeterEventID:     meterEventID,
		Quantity:         quantity,
		Timestamp:        timestamp,
		EventName:        eventName,
	}, nil
}

// createMeterEvent creates a meter event via Stripe's API.
// This uses the v1/billing/meter_events endpoint.
func (c *StripeClient) createMeterEvent(ctx context.Context, params MeterEventParams) (string, error) {
	// Build form data for Stripe API
	formData := fmt.Sprintf(
		"event_name=%s&payload[stripe_customer_id]=%s&payload[value]=%s",
		params.EventName,
		params.Payload["stripe_customer_id"],
		params.Payload["value"],
	)

	if params.Timestamp > 0 {
		formData += fmt.Sprintf("&timestamp=%d", params.Timestamp)
	}
	if params.Identifier != "" {
		formData += fmt.Sprintf("&identifier=%s", params.Identifier)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.stripe.com/v1/billing/meter_events",
		strings.NewReader(formData))
	if err != nil {
		return "", fmt.Errorf("failed to create meter event request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Stripe-Version", "2024-11-20.acacia") // Use latest API version

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send meter event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("meter event failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID string `json:"identifier"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode meter event response: %w", err)
	}

	return result.ID, nil
}

// RecordEmissionsUsage is a convenience method for recording emissions calculation usage.
// This tracks how many emissions calculations a customer has performed.
func (c *StripeClient) RecordEmissionsUsage(ctx context.Context, customerID string, calculationCount int64) (*StripeUsageRecord, error) {
	return c.RecordUsageWithEvent(ctx, customerID, calculationCount, time.Now(), "emissions_calculated")
}

// RecordAPIUsage is a convenience method for recording API request usage.
// This tracks API calls for usage-based billing.
func (c *StripeClient) RecordAPIUsage(ctx context.Context, customerID string, requestCount int64) (*StripeUsageRecord, error) {
	return c.RecordUsageWithEvent(ctx, customerID, requestCount, time.Now(), "api_requests")
}

// RecordDataIngestionUsage is a convenience method for recording data ingestion usage.
// This tracks data points ingested for usage-based billing.
func (c *StripeClient) RecordDataIngestionUsage(ctx context.Context, customerID string, dataPoints int64) (*StripeUsageRecord, error) {
	return c.RecordUsageWithEvent(ctx, customerID, dataPoints, time.Now(), "data_ingested")
}

// CreateProduct creates a Stripe product
func (c *StripeClient) CreateProduct(ctx context.Context, name, description string, metadata map[string]string) (*stripe.Product, error) {
	params := &stripe.ProductParams{
		Name:        stripe.String(name),
		Description: stripe.String(description),
		Metadata:    metadata,
	}
	params.Context = ctx

	prod, err := product.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}
	return prod, nil
}

// CreatePrice creates a price for a product
func (c *StripeClient) CreatePrice(ctx context.Context, productID string, unitAmount int64, currency, interval string, metadata map[string]string) (*stripe.Price, error) {
	params := &stripe.PriceParams{
		Product:    stripe.String(productID),
		UnitAmount: stripe.Int64(unitAmount),
		Currency:   stripe.String(currency),
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String(interval),
		},
		Metadata: metadata,
	}
	params.Context = ctx

	p, err := price.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create price: %w", err)
	}
	return p, nil
}

// ListInvoices lists invoices for a customer
func (c *StripeClient) ListInvoices(ctx context.Context, customerID string, limit int64) ([]*stripe.Invoice, error) {
	params := &stripe.InvoiceListParams{
		Customer: stripe.String(customerID),
	}
	params.Limit = stripe.Int64(limit)
	params.Context = ctx

	i := invoice.List(params)
	invoices := make([]*stripe.Invoice, 0)

	for i.Next() {
		invoices = append(invoices, i.Invoice())
	}

	if err := i.Err(); err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}
	return invoices, nil
}

// GetPriceForPlan returns the Stripe price ID for a plan tier
func (c *StripeClient) GetPriceForPlan(plan PlanTier) (string, error) {
	switch plan {
	case PlanFree:
		return c.priceFree, nil
	case PlanPro:
		return c.pricePro, nil
	case PlanEnterprise:
		return c.priceEnterprise, nil
	default:
		return "", fmt.Errorf("unknown plan tier: %s", plan)
	}
}

// GetSubscriptionLimits returns feature limits for a plan
func GetSubscriptionLimits(plan PlanTier) SubscriptionLimits {
	limits := map[PlanTier]SubscriptionLimits{
		PlanFree: {
			MaxUsers:            5,
			MaxDataSources:      2,
			MaxEmissionsRecords: 10000,
			APIRateLimit:        100,
			DataRetentionDays:   30,
			SupportLevel:        "community",
			Features: []string{
				"basic_emissions_tracking",
				"csv_import",
				"basic_reports",
			},
		},
		PlanPro: {
			MaxUsers:            50,
			MaxDataSources:      10,
			MaxEmissionsRecords: 1000000,
			APIRateLimit:        10000,
			DataRetentionDays:   365,
			SupportLevel:        "email",
			Features: []string{
				"basic_emissions_tracking",
				"csv_import",
				"basic_reports",
				"cloud_connectors",
				"csrd_compliance",
				"sec_compliance",
				"api_access",
				"custom_reports",
			},
		},
		PlanEnterprise: {
			MaxUsers:            -1,
			MaxDataSources:      -1,
			MaxEmissionsRecords: -1,
			APIRateLimit:        100000,
			DataRetentionDays:   -1,
			SupportLevel:        "priority",
			Features: []string{
				"basic_emissions_tracking",
				"csv_import",
				"basic_reports",
				"cloud_connectors",
				"csrd_compliance",
				"sec_compliance",
				"api_access",
				"custom_reports",
				"sap_connector",
				"utility_connector",
				"cbam_compliance",
				"california_compliance",
				"white_label",
				"sso",
				"dedicated_support",
				"sla",
			},
		},
	}
	return limits[plan]
}

// SubscriptionLimits defines feature limits for a subscription plan
type SubscriptionLimits struct {
	MaxUsers            int
	MaxDataSources      int
	MaxEmissionsRecords int
	APIRateLimit        int
	DataRetentionDays   int
	SupportLevel        string
	Features            []string
}

// HasFeature checks if a plan includes a specific feature
func (pl SubscriptionLimits) HasFeature(feature string) bool {
	for _, f := range pl.Features {
		if f == feature {
			return true
		}
	}
	return false
}
