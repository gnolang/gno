package pager

import (
	"math"
	"net/url"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

// Pager is a struct that holds the AVL tree and pagination parameters.
type Pager struct {
	Tree            *avl.Tree
	PageQueryParam  string
	DefaultPageSize int
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
}

// Item represents a key-value pair in the AVL tree.
type Item struct {
	Key   string
	Value interface{}
}

// NewPager creates a new Pager with default values.
func NewPager(tree *avl.Tree) *Pager {
	return &Pager{
		Tree:            tree,
		PageQueryParam:  "page",
		DefaultPageSize: 10,
	}
}

// GetPage retrieves a page of results from the AVL tree.
func (p *Pager) GetPage(pageNumber int, pageSize int) (*Page, error) {
	if pageSize < 1 {
		return nil, ufmt.Errorf("invalid page size")
	}

	totalItems := p.Tree.Size()
	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))

	if pageNumber < 1 {
		return &Page{
			Items:      []Item{},
			PageNumber: 0,
			PageSize:   pageSize,
			TotalItems: totalItems,
			TotalPages: totalPages,
			HasPrev:    false,
			HasNext:    pageNumber < totalPages,
		}, nil
	}

	if pageNumber > totalPages {
		return &Page{
			Items:      []Item{},
			PageNumber: pageNumber,
			PageSize:   pageSize,
			TotalItems: totalItems,
			TotalPages: totalPages,
			HasPrev:    totalPages > 0,
			HasNext:    false,
		}, nil
	}

	startIndex := (pageNumber - 1) * pageSize
	endIndex := startIndex + pageSize
	if endIndex > totalItems {
		endIndex = totalItems
	}

	items := []Item{}
	p.Tree.ReverseIterateByOffset(startIndex, endIndex-startIndex, func(key string, value interface{}) bool {
		items = append(items, Item{Key: key, Value: value})
		return false
	})

	return &Page{
		Items:      items,
		PageNumber: pageNumber,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasPrev:    pageNumber > 1,
		HasNext:    pageNumber < totalPages,
	}, nil
}

// RenderSelector generates the Markdown for the page selector.
func (p *Page) RenderSelector(queryParam string) string {
	if p.TotalPages <= 1 {
		return ""
	}

	md := ""

	if p.HasPrev {
		// Always show the first page link
		md += ufmt.Sprintf("[%d](?%s=%d) | ", 1, queryParam, 1)

		// Before
		if p.PageNumber == 5 {
			md += ufmt.Sprintf("[%d](?%s=%d) | ", 2, queryParam, 2)
		} else if p.PageNumber > 4 {
			md += "… | "
		}

		if p.PageNumber > 3 {
			md += ufmt.Sprintf("[%d](?%s=%d) | ", p.PageNumber-2, queryParam, p.PageNumber-2)
		}

		if p.PageNumber > 2 {
			md += ufmt.Sprintf("[%d](?%s=%d) | ", p.PageNumber-1, queryParam, p.PageNumber-1)
		}
	}

	// Current page
	md += ufmt.Sprintf("**%d**", p.PageNumber)

	if p.HasNext {
		md += " | "

		if p.PageNumber < p.TotalPages-1 {
			md += ufmt.Sprintf("[%d](?%s=%d) | ", p.PageNumber+1, queryParam, p.PageNumber+1)
		}

		if p.PageNumber < p.TotalPages-2 {
			md += ufmt.Sprintf("[%d](?%s=%d) | ", p.PageNumber+2, queryParam, p.PageNumber+2)
		}

		if p.PageNumber == p.TotalPages-4 {
			md += ufmt.Sprintf("[%d](?%s=%d) | ", p.PageNumber+3, queryParam, p.PageNumber+3)
		} else if p.PageNumber < p.TotalPages-3 {
			md += "… | "
		}

		// Always show the last page link
		md += ufmt.Sprintf("[%d](?%s=%d)", p.TotalPages, queryParam, p.TotalPages)
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

	if pageStr := query.Get(p.PageQueryParam); pageStr != "" {
		pageNumber, err = strconv.Atoi(pageStr)
		if err != nil || pageNumber < 1 {
			pageNumber = 1
		}
	}

	if sizeStr := query.Get("size"); sizeStr != "" {
		pageSize, err = strconv.Atoi(sizeStr)
		if err != nil || pageSize < 1 {
			pageSize = p.DefaultPageSize
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
