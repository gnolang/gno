package pager

import (
	"math"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

// Pager is a struct that holds the AVL tree and pagination parameters.
type Pager struct {
	Tree *avl.Tree
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
func (p *Page) RenderSelector() string {
	if p.TotalPages <= 1 {
		return ""
	}

	md := ""

	// First and Previous links
	if p.HasPrev {
		md += ufmt.Sprintf("[First](?page=1) | ")
		md += ufmt.Sprintf("[Previous](?page=%d) | ", p.PageNumber-1)
	} else {
		md += "First | Previous | "
	}

	// Page number links
	startPage := max(1, p.PageNumber-2)
	endPage := min(p.TotalPages, p.PageNumber+2)

	if startPage > 1 {
		md += "… | "
	}

	for i := startPage; i <= endPage; i++ {
		if i == p.PageNumber {
			md += ufmt.Sprintf("**%d** | ", i)
		} else {
			md += ufmt.Sprintf("[%d](?page=%d) | ", i, i)
		}
	}

	if endPage < p.TotalPages {
		md += "… | "
	}

	// Next and Last links
	if p.HasNext {
		md += ufmt.Sprintf("[Next](?page=%d) | ", p.PageNumber+1)
		md += ufmt.Sprintf("[Last](?page=%d)", p.TotalPages)
	} else {
		md += "Next | Last"
	}

	return md
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
