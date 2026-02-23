// Package pageable implements a flexible pagination system that can be used with any data structure
// that implements the Pageable interface. It provides functionality for:
//
//   - Paginating through collections of items
//   - Configurable page sizes
//   - Forward and reverse iteration
//   - URL query parameter parsing
//   - Markdown-based pagination UI
//
// This package is currently used by:
//   - gno.land/p/nt/avl/pager: implementation for AVL trees.
//   - gno.land/p/moul/ulist/pager: implementation for ulist
package pageable

import (
	"math"
	"net/url"
	"strconv"

	"gno.land/p/nt/ufmt"
)

// Pageable defines the minimal interface required for pagination.
// This interface is intentionally lightweight, requiring only two methods.
// While some data structures (like AVL trees) might have separate methods for forward
// and reverse iteration, this interface consolidates both directions into a single
// IterateByOffset method. It's the responsibility of the implementing wrapper to
// handle the direction logic internally based on the offset and count parameters.
type Pageable interface {
	// Size returns the total number of items
	Size() int

	// IterateByOffset performs iteration starting from offset for count elements.
	// The callback receives an index and a value, returns true to stop iteration.
	//
	// For reverse iteration, the implementation should handle the direction internally
	// by adjusting how it interprets the offset and count parameters. For example:
	// - In forward mode: offset 0, count 5 would return items [0,1,2,3,4]
	// - In reverse mode: offset 0, count -5 would return items [4,3,2,1,0]
	//
	// This approach allows the interface to remain simple while supporting
	// bidirectional iteration through the implementation layer.
	IterateByOffset(offset int, count int, cb func(index interface{}, value interface{}) bool) bool
}

// Pager provides pagination functionality for any Pageable source
type Pager struct {
	Source          Pageable
	PageQueryParam  string
	SizeQueryParam  string
	DefaultPageSize int
	Reversed        bool
}

// Page represents a single page of results
type Page struct {
	Items      []Item
	PageNumber int
	PageSize   int
	TotalItems int
	TotalPages int
	HasPrev    bool
	HasNext    bool
	Pager      *Pager
}

// Item represents a generic item in the page
type Item struct {
	Index interface{}
	Value interface{}
}

// NewPager creates a new Pager with default values
func NewPager(source Pageable, defaultPageSize int, reversed bool) *Pager {
	return &Pager{
		Source:          source,
		PageQueryParam:  "page",
		SizeQueryParam:  "size",
		DefaultPageSize: defaultPageSize,
		Reversed:        reversed,
	}
}

// GetPage retrieves a page of results from the AVL tree.
func (p *Pager) GetPage(pageNumber int) *Page {
	return p.GetPageWithSize(pageNumber, p.DefaultPageSize)
}

// Helper function to create a basic page
func (p *Pager) newBasePage(pageSize int) *Page {
	totalItems := p.Source.Size()
	totalPages := 0
	if pageSize > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(pageSize)))
	}

	return &Page{
		TotalItems: totalItems,
		TotalPages: totalPages,
		PageSize:   pageSize,
		Pager:      p,
	}
}

func (p *Pager) GetPageWithSize(pageNumber, pageSize int) *Page {
	// Handle invalid page size
	if pageSize <= 0 {
		return p.newBasePage(pageSize)
	}

	page := p.newBasePage(pageSize)

	// For empty source, return empty page with no navigation
	if page.TotalItems == 0 {
		page.HasPrev = false
		page.HasNext = false
		return page
	}

	// page number provided is not available
	if pageNumber < 1 {
		page.HasNext = page.TotalPages > 0
		return page
	}

	// page number provided is outside the range of total pages
	if pageNumber > page.TotalPages {
		page.PageNumber = pageNumber
		page.HasPrev = page.TotalPages > 0
		return page
	}

	// Calculate offset and size
	var offset, iterCount int
	if p.Reversed {
		// For reversed order, calculate from the end
		iterCount = min(pageSize, page.TotalItems-((pageNumber-1)*pageSize))
		offset = max(0, page.TotalItems-(pageNumber*pageSize))
	} else {
		offset = (pageNumber - 1) * pageSize
		iterCount = pageSize
	}

	// Collect items for the current page
	var items []Item
	if p.Reversed {
		// For reversed order, use negative count to indicate reverse iteration
		offset = page.TotalItems - offset - iterCount
		p.Source.IterateByOffset(offset, -iterCount, func(index interface{}, value interface{}) bool {
			items = append(items, Item{Index: index, Value: value})
			return false
		})
	} else {
		p.Source.IterateByOffset(offset, iterCount, func(index interface{}, value interface{}) bool {
			items = append(items, Item{Index: index, Value: value})
			return false
		})
	}

	page.Items = items
	page.PageNumber = pageNumber
	page.HasPrev = pageNumber > 1
	page.HasNext = pageNumber < page.TotalPages
	return page
}

func (p *Pager) MustGetPageByPath(rawURL string) *Page {
	page, err := p.GetPageByPath(rawURL)
	if err != nil {
		panic("invalid path")
	}
	return page
}

// GetPageByPath retrieves a page of results based on the query parameters in the URL path.
func (p *Pager) GetPageByPath(rawURL string) (*Page, error) {
	pageNumber, pageSize, err := p.ParseQuery(rawURL)
	if err != nil {
		return nil, err
	}
	return p.GetPageWithSize(pageNumber, pageSize), nil
}

// Picker generates the Markdown UI for the page picker using default query string format
func (p *Page) Picker() string {
	return p.PickerWithPath("")
}

// PickerWithPath generates the Markdown UI for the page picker with optional path
// If path is provided, it preserves existing query parameters (except page)
func (p *Page) PickerWithPath(path string) string {
	pageNumber := p.PageNumber
	pageNumber = max(pageNumber, 1)

	if p.TotalPages <= 1 {
		return ""
	}

	// Parse existing query parameters if path is provided
	baseQuery := ""
	if path != "" {
		u, _ := url.Parse(path)
		if u != nil {
			query := u.Query()
			// Remove existing page query parameter
			query.Del(p.Pager.PageQueryParam)
			// Encode remaining query parameters
			baseQuery = query.Encode()
			if baseQuery != "" {
				baseQuery = "&" + baseQuery
			}
		}
	}

	md := ""

	if p.HasPrev {
		// Always show the first page link
		md += ufmt.Sprintf("[%d](?%s=%d%s) | ", 1, p.Pager.PageQueryParam, 1, baseQuery)

		// Before
		if p.PageNumber > 4 {
			md += "… | "
		}

		if p.PageNumber > 3 {
			md += ufmt.Sprintf("[%d](?%s=%d%s) | ", p.PageNumber-2, p.Pager.PageQueryParam, p.PageNumber-2, baseQuery)
		}

		if p.PageNumber > 2 {
			md += ufmt.Sprintf("[%d](?%s=%d%s) | ", p.PageNumber-1, p.Pager.PageQueryParam, p.PageNumber-1, baseQuery)
		}
	}

	if p.PageNumber > 0 && p.PageNumber <= p.TotalPages {
		// Current page
		md += ufmt.Sprintf("**%d**", p.PageNumber)
	} else {
		md += ufmt.Sprintf("_%d_", p.PageNumber)
	}

	if p.HasNext {
		md += " | "

		if p.PageNumber < p.TotalPages-1 {
			md += ufmt.Sprintf("[%d](?%s=%d%s) | ", p.PageNumber+1, p.Pager.PageQueryParam, p.PageNumber+1, baseQuery)
		}

		if p.PageNumber < p.TotalPages-2 {
			md += ufmt.Sprintf("[%d](?%s=%d%s) | ", p.PageNumber+2, p.Pager.PageQueryParam, p.PageNumber+2, baseQuery)
		}

		if p.PageNumber < p.TotalPages-3 {
			md += "… | "
		}

		// Always show the last page link
		md += ufmt.Sprintf("[%d](?%s=%d%s)", p.TotalPages, p.Pager.PageQueryParam, p.TotalPages, baseQuery)
	}

	return md
}

// Helper to safely parse positive integers with a default value
func parsePositiveIntOrDefault(str string, defaultValue int) int {
	if str == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(str)
	if err != nil || val < 1 {
		return defaultValue
	}
	return val
}

func (p *Pager) ParseQuery(rawURL string) (int, int, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return 1, p.DefaultPageSize, err
	}

	query := u.Query()
	pageNumber := parsePositiveIntOrDefault(query.Get(p.PageQueryParam), 1)
	pageSize := parsePositiveIntOrDefault(query.Get(p.SizeQueryParam), p.DefaultPageSize)

	return pageNumber, pageSize, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
