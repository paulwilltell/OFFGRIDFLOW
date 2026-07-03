package billing

import (
	"context"
	"testing"
)

func TestReportEntitlement_FailsClosedThenGrants(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryReportStore()
	svc := &Service{reportStore: store}

	const tenant = "tenant-A"

	// Before any purchase, the tenant must NOT be entitled.
	if svc.HasPaidForReport(ctx, tenant) {
		t.Fatal("tenant should not have report access before paying")
	}

	// Record a purchase (as the webhook would on checkout.session.completed).
	if err := svc.RecordReportPurchase(ctx, tenant, "cs_test_123"); err != nil {
		t.Fatalf("record purchase failed: %v", err)
	}

	// Now entitled.
	if !svc.HasPaidForReport(ctx, tenant) {
		t.Fatal("tenant should have report access after paying")
	}

	// A different tenant remains locked (no cross-tenant leakage).
	if svc.HasPaidForReport(ctx, "tenant-B") {
		t.Fatal("unrelated tenant must not inherit report access")
	}
}

func TestReportEntitlement_Idempotent(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryReportStore()

	p := &ReportPurchase{TenantID: "t1", StripeSessionID: "cs_1"}
	if err := store.Record(ctx, p); err != nil {
		t.Fatalf("first record failed: %v", err)
	}
	// Same session ID again — must not error (webhooks can be delivered twice).
	if err := store.Record(ctx, p); err != nil {
		t.Fatalf("duplicate record should be idempotent, got: %v", err)
	}

	paid, err := store.HasPaid(ctx, "t1")
	if err != nil || !paid {
		t.Fatalf("expected tenant paid=true, got paid=%v err=%v", paid, err)
	}
}

func TestHasPaidForReport_FailsClosedWithoutStore(t *testing.T) {
	// A service with no report store must treat everyone as unpaid.
	svc := &Service{}
	if svc.HasPaidForReport(context.Background(), "any") {
		t.Fatal("service without report store must fail closed (unpaid)")
	}
}
