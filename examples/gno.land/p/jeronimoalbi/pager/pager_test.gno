package pager

import (
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
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
			uri:       "gno.land/r/devxteam/gnome:foo/bar?page=1&foo=bar",
			items:     []int{1, 2, 3, 4, 5, 6},
			hasPages:  true,
			nextPath:  "?foo=bar&page=2",
			pageSize:  5,
			page:      1,
			pageCount: 2,
		},
		{
			name:      "page 2",
			uri:       "gno.land/r/devxteam/gnome:foo/bar?page=2&foo=bar",
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
			uri:       "gno.land/r/devxteam/gnome:foo/bar?current=2&foo=bar",
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
			uri:  "gno.land/r/devxteam/gnome:foo/bar?page=3&foo=bar",
			err:  ErrInvalidPageNumber,
		},
		{
			name: "invalid page zero",
			uri:  "gno.land/r/devxteam/gnome:foo/bar?page=0",
			err:  ErrInvalidPageNumber,
		},
		{
			name: "invalid page number",
			uri:  "gno.land/r/devxteam/gnome:foo/bar?page=foo",
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
		name, uri     string
		items, page   []int
		stop, stopped bool
	}{
		{
			name:  "page 1",
			uri:   "gno.land/r/devxteam/gnome:foo/bar?page=1",
			items: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
			page:  []int{1, 2, 3},
		},
		{
			name:  "page 2",
			uri:   "gno.land/r/devxteam/gnome:foo/bar?page=2",
			items: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
			page:  []int{4, 5, 6},
		},
		{
			name:  "page 3",
			uri:   "gno.land/r/devxteam/gnome:foo/bar?page=3",
			items: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
			page:  []int{7, 8, 9},
		},
		{
			name:    "invalid page",
			uri:     "gno.land/r/devxteam/gnome:foo/bar?page=2",
			items:   []int{1, 2, 3},
			stopped: true,
		},
		{
			name:    "stop iteration",
			uri:     "gno.land/r/devxteam/gnome:foo/bar?page=1",
			items:   []int{1, 2, 3},
			stop:    true,
			stopped: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			var (
				items []int
				p, _  = New(tc.uri, len(tc.items), WithPageSize(3))
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
			uassert.Equal(t, tc.stopped, stopped)
			urequire.Equal(t, len(tc.page), len(items), "expect iteration of the right number of items")

			for i, v := range items {
				urequire.Equal(t, tc.page[i], v, "expect iterated items to match")
			}
		})
	}
}
