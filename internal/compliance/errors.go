package compliance

import "errors"

var (
	// Validation errors
	ErrInvalidTenant     = errors.New("invalid tenant ID")
	ErrInvalidUser       = errors.New("invalid user ID")
	ErrInvalidYear       = errors.New("invalid reporting year")
	ErrInvalidPeriod     = errors.New("invalid reporting period")
	ErrInvalidReportType = errors.New("invalid report type")
	
	// Report errors
	ErrReportNotFound    = errors.New("compliance report not found")
	ErrReportExists      = errors.New("report already exists for this period")
	ErrInsufficientData  = errors.New("insufficient emissions data for report")
	
	// Generation errors
	ErrPDFGeneration     = errors.New("PDF generation failed")
	ErrXBRLGeneration    = errors.New("XBRL generation failed")
	ErrDataQualityLow    = errors.New("data quality score too low for compliance report")
)
