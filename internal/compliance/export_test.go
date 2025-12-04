package compliance

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExportSummaryHelpers(t *testing.T) {
	summary := &ComplianceSummary{
		Frameworks: map[string]FrameworkSummary{
			"csrd": {
				Name:   "CSRD/ESRS E1",
				Status: StatusPartial,
				Scope1: true,
				Scope2: true,
				Scope3: false,
			},
		},
		Totals: EmissionsTotals{
			Scope1Tons: 1.25,
			Scope2Tons: 0.75,
			Scope3Tons: 0.5,
			TotalTons:  2.5,
		},
	}

	pdfBytes, err := ExportSummaryToPDF(summary)
	require.NoError(t, err)
	require.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF")), "pdf should start with %PDF")

	xbrlBytes, err := ExportSummaryToXBRL(summary)
	require.NoError(t, err)
	require.Contains(t, string(xbrlBytes), "<framework>")
	require.Contains(t, string(xbrlBytes), "<scope1Tons>")
}
