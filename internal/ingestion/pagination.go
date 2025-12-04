// Package ingestion provides pagination helpers for cloud APIs.
package ingestion

import (
	"fmt"
)

// PaginationState tracks pagination across API calls.
// Supports both cursor-based (nextToken) and offset-based pagination.
type PaginationState struct {
	// Cursor for cursor-based pagination (e.g., "nextToken", "continuation_token")
	Cursor string

	// Offset for offset-based pagination
	Offset int

	// PageSize is the number of items per page
	PageSize int

	// MaxPages limits total pages fetched (0 = unlimited)
	MaxPages int

	// CurrentPage tracks pages fetched so far
	CurrentPage int

	// TotalFetched tracks total items fetched across all pages
	TotalFetched int

	// HasMore indicates if more data is available
	HasMore bool
}

// NewPaginationState creates a new pagination state with default page size.
func NewPaginationState(pageSize int) *PaginationState {
	if pageSize <= 0 {
		pageSize = 100
	}
	if pageSize > 10000 {
		pageSize = 10000 // Cap at 10k to prevent API abuse
	}

	return &PaginationState{
		PageSize:    pageSize,
		CurrentPage: 0,
		TotalFetched: 0,
		HasMore:     true,
	}
}

// SetCursorBased initializes for cursor-based pagination.
func (p *PaginationState) SetCursorBased(cursor string) {
	p.Cursor = cursor
	p.HasMore = cursor != ""
}

// SetOffsetBased initializes for offset-based pagination.
func (p *PaginationState) SetOffsetBased(offset int) {
	p.Offset = offset
	p.HasMore = offset >= 0
}

// NextOffset returns the offset for the next page request.
func (p *PaginationState) NextOffset() int {
	return p.Offset + (p.CurrentPage * p.PageSize)
}

// AdvancePage moves to the next page and returns error if max pages exceeded.
func (p *PaginationState) AdvancePage() error {
	// Check max pages limit
	if p.MaxPages > 0 && p.CurrentPage >= p.MaxPages {
		p.HasMore = false
		return fmt.Errorf("pagination: reached max pages limit (%d)", p.MaxPages)
	}

	p.CurrentPage++
	return nil
}

// UpdateCursor updates cursor for next request and tracks if more data available.
func (p *PaginationState) UpdateCursor(newCursor string, itemCount int) error {
	p.TotalFetched += itemCount
	p.Cursor = newCursor
	p.HasMore = newCursor != "" && (p.MaxPages == 0 || p.CurrentPage < p.MaxPages)

	return p.AdvancePage()
}

// IsDone returns true if pagination is complete.
func (p *PaginationState) IsDone() bool {
	return !p.HasMore || (p.MaxPages > 0 && p.CurrentPage >= p.MaxPages)
}

// Summary returns a human-readable summary of pagination progress.
func (p *PaginationState) Summary() string {
	return fmt.Sprintf("pages=%d items=%d hasMore=%v", p.CurrentPage, p.TotalFetched, p.HasMore)
}

// PageRange represents a range of items in a page.
type PageRange struct {
	Offset int // Starting index
	Limit  int // Number of items
}

// NewPageRange creates a page range for offset-based pagination.
func NewPageRange(pageNum, pageSize int) *PageRange {
	if pageNum < 0 {
		pageNum = 0
	}
	if pageSize <= 0 {
		pageSize = 100
	}

	return &PageRange{
		Offset: pageNum * pageSize,
		Limit:  pageSize,
	}
}

// NextRange returns the range for the next page.
func (pr *PageRange) NextRange() *PageRange {
	return NewPageRange((pr.Offset/pr.Limit)+1, pr.Limit)
}

// String returns a string representation of the range.
func (pr *PageRange) String() string {
	return fmt.Sprintf("offset=%d limit=%d", pr.Offset, pr.Limit)
}
