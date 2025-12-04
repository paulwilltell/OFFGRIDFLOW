// Package aws provides test data helpers for integration tests.
package aws

import (
	"fmt"
)

// SampleCURRecords returns sample Cost and Usage Report records for testing.
func SampleCURRecords(count int) []struct {
	Date        string
	Service     string
	UsageAmount float64
	Cost        float64
} {
	records := make([]struct {
		Date        string
		Service     string
		UsageAmount float64
		Cost        float64
	}, count)

	for i := 0; i < count; i++ {
		records[i] = struct {
			Date        string
			Service     string
			UsageAmount float64
			Cost        float64
		}{
			Date:        fmt.Sprintf("2024-01-%02d", (i%28)+1),
			Service:     "AmazonEC2",
			UsageAmount: float64(i+1) * 100.0,
			Cost:        float64(i+1) * 50.0,
		}
	}

	return records
}
