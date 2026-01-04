// Package pager provides pagination functionality through a generic pager implementation.
//
// Example usage:
//
//	import (
//	    "strconv"
//	    "strings"
//
//	    "gno.land/p/jeronimoalbi/pager"
//	)
//
//	func Render(path string) string {
//	    // Define the items to paginate
//	    items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
//
//	    // Create a pager that paginates 4 items at a time
//	    p, err := pager.New(path, len(items), pager.WithPageSize(4))
//	    if err != nil {
//	        panic(err)
//	    }
//
//	    // Render items for the current page
//	    var output strings.Builder
//	    p.Iterate(func(i int) bool {
//	        output.WriteString("- " + strconv.Itoa(items[i]) + "\n")
//	        return false
//	    })
//
//	    // Render page picker
//	    if p.HasPages() {
//	        output.WriteString("\n" + pager.Picker(p))
//	    }
//
//	    return output.String()
//	}
package pager

import (
	"errors"
	"math"
	"net/url"
	"strconv"
	"strings"
)

var ErrInvalidPageNumber = errors.New("invalid page number")

// PagerIterFn defines a callback to iterate page items.
type PagerIterFn func(index int) (stop bool)

// New creates a new pager.
func New(rawURL string, totalItems int, options ...PagerOption) (Pager, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return Pager{}, err
	}

	p := Pager{
		query:          u.RawQuery,
		pageQueryParam: DefaultPageQueryParam,
		pageSize:       DefaultPageSize,
		page:           1,
		totalItems:     totalItems,
	}
	for _, apply := range options {
		apply(&p)
	}

	p.pageCount = int(math.Ceil(float64(p.totalItems) / float64(p.pageSize)))

	rawPage := u.Query().Get(p.pageQueryParam)
	if rawPage != "" {
		p.page, _ = strconv.Atoi(rawPage)
		if p.page == 0 || p.page > p.pageCount {
			return Pager{}, ErrInvalidPageNumber
		}
	}

	return p, nil
}

// MustNew creates a new pager or panics if there is an error.
func MustNew(rawURL string, totalItems int, options ...PagerOption) Pager {
	p, err := New(rawURL, totalItems, options...)
	if err != nil {
		panic(err)
	}
	return p
}

// Pager allows paging items.
type Pager struct {
	query, pageQueryParam                 string
	pageSize, page, pageCount, totalItems int
}

// TotalItems returns the total number of items to paginate.
func (p Pager) TotalItems() int {
	return p.totalItems
}

// PageSize returns the size of each page.
func (p Pager) PageSize() int {
	return p.pageSize
}

// Page returns the current page number.
func (p Pager) Page() int {
	return p.page
}

// PageCount returns the number pages.
func (p Pager) PageCount() int {
	return p.pageCount
}

// Offset returns the index of the first page item.
func (p Pager) Offset() int {
	return (p.page - 1) * p.pageSize
}

// HasPages checks if pager has more than one page.
func (p Pager) HasPages() bool {
	return p.pageCount > 1
}

// GetPageURI returns the URI for a page.
// An empty string is returned when page doesn't exist.
func (p Pager) GetPageURI(page int) string {
	if page < 1 || page > p.PageCount() {
		return ""
	}

	values, _ := url.ParseQuery(p.query)
	values.Set(p.pageQueryParam, strconv.Itoa(page))
	return "?" + values.Encode()
}

// PrevPageURI returns the URI path to the previous page.
// An empty string is returned when current page is the first page.
func (p Pager) PrevPageURI() string {
	if p.page == 1 || !p.HasPages() {
		return ""
	}
	return p.GetPageURI(p.page - 1)
}

// NextPageURI returns the URI path to the next page.
// An empty string is returned when current page is the last page.
func (p Pager) NextPageURI() string {
	if p.page == p.pageCount {
		// Current page is the last page
		return ""
	}
	return p.GetPageURI(p.page + 1)
}

// Iterate allows iterating page items.
func (p Pager) Iterate(fn PagerIterFn) bool {
	if p.totalItems == 0 {
		return true
	}

	start := p.Offset()
	end := start + p.PageSize()
	if end > p.totalItems {
		end = p.totalItems
	}

	for i := start; i < end; i++ {
		if fn(i) {
			return true
		}
	}
	return false
}

// TODO: Support different types of pickers (ex. with clickable page numbers)

// Picker returns a string with the pager as Markdown.
// An empty string is returned when the pager has no pages.
func Picker(p Pager) string {
	if !p.HasPages() {
		return ""
	}

	var out strings.Builder

	if s := p.PrevPageURI(); s != "" {
		out.WriteString("[«](" + s + ") | ")
	} else {
		out.WriteString("\\- | ")
	}

	out.WriteString("page " + strconv.Itoa(p.Page()) + " of " + strconv.Itoa(p.PageCount()))

	if s := p.NextPageURI(); s != "" {
		out.WriteString(" | [»](" + s + ")")
	} else {
		out.WriteString(" | \\-")
	}

	return out.String()
}
