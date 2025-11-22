package pager

import (
	"math"
	"net/url"
	"strconv"

	"gno.land/p/nt/avl/rotree"
	"gno.land/p/nt/ufmt"
)

// Pager is a struct that holds the AVL tree and pagination parameters.
type Pager struct {
	Tree            rotree.IReadOnlyTree
	PageQueryParam  string
	SizeQueryParam  string
	DefaultPageSize int
	Reversed        bool
}

// Page represents a single page of results.
type Page struct {
	Items      []Item
	PageNumber int
	PageSize   int
	TotalItems int
	TotalPages int
	HasPrev    bool
	HasNext    bool
	Pager      *Pager // Reference to the parent Pager
}

// Item represents a key-value pair in the AVL tree.
type Item struct {
	Key   string
	Value any
}

// NewPager creates a new Pager with default values.
func NewPager(tree rotree.IReadOnlyTree, defaultPageSize int, reversed bool) *Pager {
	return &Pager{
		Tree:            tree,
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

func (p *Pager) GetPageWithSize(pageNumber, pageSize int) *Page {
	if pageNumber < 0 || pageSize < 0 {
		panic("GetPageWithSize: invalid page number or page size")
	}

	totalItems := p.Tree.Size()
	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))

	page := &Page{
		TotalItems: totalItems,
		TotalPages: totalPages,
		PageSize:   pageSize,
		Pager:      p,
	}

	// pages without content
	if pageSize < 1 {
		return page
	}

	// page number provided is not available
	if pageNumber < 1 {
		page.HasNext = totalPages > 0
		return page
	}

	// page number provided is outside the range of total pages
	if pageNumber > totalPages {
		page.PageNumber = pageNumber
		page.HasPrev = pageNumber > 0
		return page
	}

	startIndex := (pageNumber - 1) * pageSize
	endIndex := startIndex + pageSize
	if endIndex > totalItems {
		endIndex = totalItems
	}

	items := []Item{}

	if p.Reversed {
		p.Tree.ReverseIterateByOffset(startIndex, endIndex-startIndex, func(key string, value any) bool {
			items = append(items, Item{Key: key, Value: value})
			return false
		})
	} else {
		p.Tree.IterateByOffset(startIndex, endIndex-startIndex, func(key string, value any) bool {
			items = append(items, Item{Key: key, Value: value})
			return false
		})
	}

	page.Items = items
	page.PageNumber = pageNumber
	page.HasPrev = pageNumber > 1
	page.HasNext = pageNumber < totalPages
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

// Picker generates the Markdown UI for the page Picker
func (p *Page) Picker(path string) string {
	pageNumber := p.PageNumber
	pageNumber = max(pageNumber, 1)

	if p.TotalPages <= 1 {
		return ""
	}

	u, _ := url.Parse(path)
	query := u.Query()

	// Remove existing page query parameter
	query.Del(p.Pager.PageQueryParam)

	// Encode remaining query parameters
	baseQuery := query.Encode()
	if baseQuery != "" {
		baseQuery = "&" + baseQuery
	}
	md := ""

	if p.HasPrev {
		md += ufmt.Sprintf("[%d](?%s=%d%s) | ", 1, p.Pager.PageQueryParam, 1, baseQuery)

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
		md += ufmt.Sprintf("**%d**", p.PageNumber)
	} else {
		md += ufmt.Sprintf("_%d_", p.PageNumber)
	}

	if p.HasNext {
		if p.PageNumber < p.TotalPages-1 {
			md += ufmt.Sprintf(" | [%d](?%s=%d%s)", p.PageNumber+1, p.Pager.PageQueryParam, p.PageNumber+1, baseQuery)
		}

		if p.PageNumber < p.TotalPages-2 {
			md += ufmt.Sprintf(" | [%d](?%s=%d%s)", p.PageNumber+2, p.Pager.PageQueryParam, p.PageNumber+2, baseQuery)
		}

		if p.PageNumber < p.TotalPages-3 {
			md += " | …"
		}

		md += ufmt.Sprintf(" | [%d](?%s=%d%s)", p.TotalPages, p.Pager.PageQueryParam, p.TotalPages, baseQuery)
	}

	return md
}

// ParseQuery parses the URL to extract the page number and page size.
func (p *Pager) ParseQuery(rawURL string) (int, int, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return 1, p.DefaultPageSize, err
	}

	query := u.Query()
	pageNumber := 1
	pageSize := p.DefaultPageSize

	if p.PageQueryParam != "" {
		if pageStr := query.Get(p.PageQueryParam); pageStr != "" {
			pageNumber, err = strconv.Atoi(pageStr)
			if err != nil || pageNumber < 1 {
				pageNumber = 1
			}
		}
	}

	if p.SizeQueryParam != "" {
		if sizeStr := query.Get(p.SizeQueryParam); sizeStr != "" {
			pageSize, err = strconv.Atoi(sizeStr)
			if err != nil || pageSize < 1 {
				pageSize = p.DefaultPageSize
			}
		}
	}

	return pageNumber, pageSize, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
