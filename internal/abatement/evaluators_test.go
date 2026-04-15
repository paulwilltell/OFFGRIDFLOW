package abatement

import "testing"

func TestEvaluateScope2MarketBased(t *testing.T) {
	tests := []struct {
		name          string
		justification string
		evidence      []EvidenceUpload
		wantStatus    EngineStatus
	}{
		{
			name:          "recommended with REC and evidence",
			justification: "Uploaded renewable energy certificates covering all office meters for the reporting year.",
			evidence:      []EvidenceUpload{{FileName: "recs.pdf", MimeType: "application/pdf", Content: []byte("pdf")}},
			wantStatus:    StatusRecommended,
		},
		{
			name:          "needs clarification without evidence",
			justification: "Uploaded RECs for all three offices.",
			evidence:      nil,
			wantStatus:    StatusNeedsClarification,
		},
		{
			name:          "insufficient without market-based instrument",
			justification: "Updated utility bills and corrected quantities.",
			evidence:      []EvidenceUpload{{FileName: "bills.pdf", MimeType: "application/pdf", Content: []byte("pdf")}},
			wantStatus:    StatusInsufficient,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := evaluateScope2MarketBased(tc.justification, tc.evidence)
			if got.Status != tc.wantStatus {
				t.Fatalf("unexpected status: got %s want %s", got.Status, tc.wantStatus)
			}
		})
	}
}

func TestEvaluateScope3SupplierAction(t *testing.T) {
	got := evaluateScope3SupplierAction(
		"Collected supplier PCF data from the vendor questionnaire and attached the export.",
		[]EvidenceUpload{{FileName: "supplier.csv", MimeType: "text/csv", Content: []byte("id,value")}},
	)
	if got.Status != StatusRecommended {
		t.Fatalf("expected recommended, got %s", got.Status)
	}
	if len(got.CriteriaChecked) < 2 {
		t.Fatalf("expected criteria checks to be recorded")
	}
}

func TestEvaluateImportedGoodsAction(t *testing.T) {
	got := evaluateImportedGoodsAction("Added customs declarations for the imported steel coils with HS code mapping.", nil)
	if got.Status != StatusNeedsClarification {
		t.Fatalf("expected needs clarification, got %s", got.Status)
	}
}

func TestFrameworkDefinitionsKnownFramework(t *testing.T) {
	defs, err := FrameworkDefinitions(FrameworkSB253)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(defs) == 0 {
		t.Fatalf("expected non-empty definitions")
	}
}
