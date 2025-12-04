package excel

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
)

// Generator exports data into a CSV-backed Excel-friendly byte stream.
type Generator struct{}

// Generate produces a CSV representation of the provided data. Excel can open
// CSV directly, giving customers an immediate export while avoiding heavy
// dependencies. Data is expected to be either [][]string, []map[string]string,
// or any fmt.Stringer-compatible value.
func (g *Generator) Generate(ctx context.Context, data interface{}) ([]byte, error) {
	_ = ctx

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	switch v := data.(type) {
	case [][]string:
		if err := writer.WriteAll(v); err != nil {
			return nil, err
		}
	case []map[string]string:
		// derive headers from first map
		if len(v) == 0 {
			return []byte{}, nil
		}
		headers := make([]string, 0, len(v[0]))
		for k := range v[0] {
			headers = append(headers, k)
		}
		if err := writer.Write(headers); err != nil {
			return nil, err
		}
		for _, row := range v {
			record := make([]string, 0, len(headers))
			for _, h := range headers {
				record = append(record, row[h])
			}
			if err := writer.Write(record); err != nil {
				return nil, err
			}
		}
	default:
		if err := writer.Write([]string{"Report"}); err != nil {
			return nil, err
		}
		if err := writer.Write([]string{fmt.Sprint(v)}); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
