package pagination

import (
	"math"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

// Paginator is a struct that holds the AVL tree and pagination parameters.
type Paginator struct {
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
func (p *Paginator) GetPage(pageNumber int, pageSize int) (*Page, error) {
	if pageNumber < 1 || pageSize < 1 {
		return nil, ufmt.Errorf("invalid page number or page size")
	}

	totalItems := p.Tree.Size()
	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))

	if pageNumber > totalPages {
		return &Page{
			Items:      []Item{},
			PageNumber: pageNumber,
			PageSize:   pageSize,
			TotalItems: totalItems,
			TotalPages: totalPages,
			HasPrev:    pageNumber > 1,
			HasNext:    pageNumber < totalPages,
		}, nil
	}

	startIndex := (pageNumber - 1) * pageSize
	endIndex := startIndex + pageSize
	if endIndex > totalItems {
		endIndex = totalItems
	}

	items := []Item{}
	p.Tree.IterateByOffset(startIndex, endIndex-startIndex, func(key string, value interface{}) bool {
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

// RenderPageSelector generates the Markdown for the page selector.
func (p *Paginator) RenderPageSelector(page *Page) string {
	if page.TotalPages <= 1 {
		return ""
	}

	md := ""

	// First and Previous links
	if page.HasPrev {
		md += ufmt.Sprintf("[First](?page=1) | ")
		md += ufmt.Sprintf("[Previous](?page=%d) | ", page.PageNumber-1)
	} else {
		md += "First | Previous | "
	}

	// Page number links
	startPage := max(1, page.PageNumber-2)
	endPage := min(page.TotalPages, page.PageNumber+2)

	if startPage > 1 {
		md += "… | "
	}

	for i := startPage; i <= endPage; i++ {
		if i == page.PageNumber {
			md += ufmt.Sprintf("**%d** | ", i)
		} else {
			md += ufmt.Sprintf("[%d](?page=%d) | ", i, i)
		}
	}

	if endPage < page.TotalPages {
		md += "… | "
	}

	// Next and Last links
	if page.HasNext {
		md += ufmt.Sprintf("[Next](?page=%d) | ", page.PageNumber+1)
		md += ufmt.Sprintf("[Last](?page=%d)", page.TotalPages)
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
