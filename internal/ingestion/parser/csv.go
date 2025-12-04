package parser

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/example/offgridflow/internal/ingestion"
)

// UtilityBillRow represents a normalized CSV row for utility bills.
type UtilityBillRow struct {
	MeterID     string
	Location    string
	PeriodStart time.Time
	PeriodEnd   time.Time
	QuantityKWh float64
	OrgID       string
}

// CSVParser converts uploaded CSV files into Activities.
type CSVParser struct{}

// ParseUtilityBills reads a CSV stream and converts rows to activities.
// Expected header: meter_id,location,period_start,period_end,kwh,org_id(optional)
func (p CSVParser) ParseUtilityBills(ctx context.Context, r io.Reader, defaultOrg string) ([]ingestion.Activity, []ingestion.ImportError, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("read header: %w", err)
	}

	colIndex := map[string]int{}
	for i, col := range header {
		colIndex[strings.ToLower(strings.TrimSpace(col))] = i
	}

	required := []string{"meter_id", "location", "period_start", "period_end", "kwh"}
	for _, req := range required {
		if _, ok := colIndex[req]; !ok {
			return nil, nil, fmt.Errorf("missing required column %q", req)
		}
	}

	var (
		activities []ingestion.Activity
		errorsOut  []ingestion.ImportError
		line       = 1
	)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		line++
		if err != nil {
			errorsOut = append(errorsOut, ingestion.ImportError{
				Row:     line,
				Message: fmt.Sprintf("read: %v", err),
			})
			continue
		}

		get := func(key string) string {
			if idx, ok := colIndex[key]; ok && idx < len(record) {
				return strings.TrimSpace(record[idx])
			}
			return ""
		}

		start, err := parseDate(get("period_start"))
		if err != nil {
			errorsOut = append(errorsOut, ingestion.ImportError{Row: line, Field: "period_start", Message: err.Error()})
			continue
		}

		end, err := parseDate(get("period_end"))
		if err != nil {
			errorsOut = append(errorsOut, ingestion.ImportError{Row: line, Field: "period_end", Message: err.Error()})
			continue
		}

		kwh, err := parseFloat(get("kwh"))
		if err != nil {
			errorsOut = append(errorsOut, ingestion.ImportError{Row: line, Field: "kwh", Message: err.Error()})
			continue
		}

		orgID := get("org_id")
		if orgID == "" {
			orgID = defaultOrg
		}

		act := ingestion.Activity{
			ID:          fmt.Sprintf("csv-%s-%d", get("meter_id"), line),
			Source:      "utility_bill",
			Category:    "electricity",
			MeterID:     get("meter_id"),
			Location:    get("location"),
			PeriodStart: start,
			PeriodEnd:   end,
			Quantity:    kwh,
			Unit:        "kWh",
			OrgID:       orgID,
			CreatedAt:   time.Now().UTC(),
		}

		if err := act.Validate(); err != nil {
			errorsOut = append(errorsOut, ingestion.ImportError{
				Row:     line,
				Message: err.Error(),
			})
			continue
		}

		activities = append(activities, act)
	}

	return activities, errorsOut, nil
}

func parseDate(val string) (time.Time, error) {
	layouts := []string{time.RFC3339, "2006-01-02"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, val); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date %q", val)
}

func parseFloat(val string) (float64, error) {
	if val == "" {
		return 0, fmt.Errorf("empty number")
	}
	var f float64
	_, err := fmt.Sscanf(val, "%f", &f)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q", val)
	}
	return f, nil
}
