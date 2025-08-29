package pager

import (
	"testing"

	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
)

func TestPager(t *testing.T) {
	cases := []struct {
		name, uri, prevPath, nextPath, param string
		offset, pageSize, page, pageCount    int
		hasPages                             bool
		items                                []int
		err                                  error
	}{
		{
			name:      "page 1",
			uri:       "gno.land/r/demo/test:foo/bar?page=1&foo=bar",
			items:     []int{1, 2, 3, 4, 5, 6},
			hasPages:  true,
			nextPath:  "?foo=bar&page=2",
			pageSize:  5,
			page:      1,
			pageCount: 2,
		},
		{
			name:      "page 2",
			uri:       "gno.land/r/demo/test:foo/bar?page=2&foo=bar",
			items:     []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			hasPages:  true,
			prevPath:  "?foo=bar&page=1",
			nextPath:  "",
			offset:    5,
			pageSize:  5,
			page:      2,
			pageCount: 2,
		},
		{
			name:      "custom query param",
			uri:       "gno.land/r/demo/test:foo/bar?current=2&foo=bar",
			items:     []int{1, 2, 3},
			param:     "current",
			hasPages:  true,
			prevPath:  "?current=1&foo=bar",
			nextPath:  "",
			offset:    2,
			pageSize:  2,
			page:      2,
			pageCount: 2,
		},
		{
			name: "missing page",
			uri:  "gno.land/r/demo/test:foo/bar?page=3&foo=bar",
			err:  ErrInvalidPageNumber,
		},
		{
			name: "invalid page zero",
			uri:  "gno.land/r/demo/test:foo/bar?page=0",
			err:  ErrInvalidPageNumber,
		},
		{
			name: "invalid page number",
			uri:  "gno.land/r/demo/test:foo/bar?page=foo",
			err:  ErrInvalidPageNumber,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			p, err := New(tc.uri, len(tc.items), WithPageSize(tc.pageSize), WithPageQueryParam(tc.param))

			// Assert
			if tc.err != nil {
				urequire.ErrorIs(t, err, tc.err, "expected an error")
				return
			}

			urequire.NoError(t, err, "expect no error")
			uassert.Equal(t, len(tc.items), p.TotalItems(), "total items")
			uassert.Equal(t, tc.page, p.Page(), "page number")
			uassert.Equal(t, tc.pageCount, p.PageCount(), "number of pages")
			uassert.Equal(t, tc.pageSize, p.PageSize(), "page size")
			uassert.Equal(t, tc.prevPath, p.PrevPageURI(), "prev URL page")
			uassert.Equal(t, tc.nextPath, p.NextPageURI(), "next URL page")
			uassert.Equal(t, tc.hasPages, p.HasPages(), "has pages")
			uassert.Equal(t, tc.offset, p.Offset(), "item offset")
		})
	}
}

func TestPagerIterate(t *testing.T) {
	cases := []struct {
		name, uri   string
		items, page []int
		stop        bool
	}{
		{
			name:  "page 1",
			uri:   "gno.land/r/demo/test:foo/bar?page=1",
			items: []int{1, 2, 3, 4, 5, 6, 7},
			page:  []int{1, 2, 3},
		},
		{
			name:  "page 2",
			uri:   "gno.land/r/demo/test:foo/bar?page=2",
			items: []int{1, 2, 3, 4, 5, 6, 7},
			page:  []int{4, 5, 6},
		},
		{
			name:  "page 3",
			uri:   "gno.land/r/demo/test:foo/bar?page=3",
			items: []int{1, 2, 3, 4, 5, 6, 7},
			page:  []int{7},
		},
		{
			name:  "stop iteration",
			uri:   "gno.land/r/demo/test:foo/bar?page=1",
			items: []int{1, 2, 3},
			stop:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			var (
				items []int
				p     = MustNew(tc.uri, len(tc.items), WithPageSize(3))
			)

			// Act
			stopped := p.Iterate(func(i int) bool {
				if tc.stop {
					return true
				}

				items = append(items, tc.items[i])
				return false
			})

			// Assert
			uassert.Equal(t, tc.stop, stopped)
			urequire.Equal(t, len(tc.page), len(items), "expect iteration of the right number of items")

			for i, v := range items {
				urequire.Equal(t, tc.page[i], v, "expect iterated items to match")
			}
		})
	}
}

func TestPicker(t *testing.T) {
	pageSize := 3
	cases := []struct {
		name, uri, output string
		totalItems        int
	}{
		{
			name:       "one page",
			uri:        "gno.land/r/demo/test:foo/bar?page=1",
			totalItems: 3,
			output:     "",
		},
		{
			name:       "two pages",
			uri:        "gno.land/r/demo/test:foo/bar?page=1",
			totalItems: 4,
			output:     "\\- | page 1 of 2 | [»](?page=2)",
		},
		{
			name:       "three pages",
			uri:        "gno.land/r/demo/test:foo/bar?page=1",
			totalItems: 7,
			output:     "\\- | page 1 of 3 | [»](?page=2)",
		},
		{
			name:       "three pages second page",
			uri:        "gno.land/r/demo/test:foo/bar?page=2",
			totalItems: 7,
			output:     "[«](?page=1) | page 2 of 3 | [»](?page=3)",
		},
		{
			name:       "three pages third page",
			uri:        "gno.land/r/demo/test:foo/bar?page=3",
			totalItems: 7,
			output:     "[«](?page=2) | page 3 of 3 | \\-",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			p := MustNew(tc.uri, tc.totalItems, WithPageSize(pageSize))

			// Act
			output := Picker(p)

			// Assert
			uassert.Equal(t, tc.output, output)
		})
	}
}
