package datasource

import (
	"fmt"
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestNewQuery(t *testing.T) {
	cases := []struct {
		name    string
		options []QueryOption
		setup   func() Query
	}{
		{
			name: "default",
			setup: func() Query {
				return Query{Count: DefaultQueryRecords}
			},
		},
		{
			name:    "with offset",
			options: []QueryOption{WithOffset(100)},
			setup: func() Query {
				return Query{
					Offset: 100,
					Count:  DefaultQueryRecords,
				}
			},
		},
		{
			name:    "with count",
			options: []QueryOption{WithCount(10)},
			setup: func() Query {
				return Query{Count: 10}
			},
		},
		{
			name:    "with invalid count",
			options: []QueryOption{WithCount(0)},
			setup: func() Query {
				return Query{Count: DefaultQueryRecords}
			},
		},
		{
			name:    "by tag",
			options: []QueryOption{ByTag("foo")},
			setup: func() Query {
				return Query{
					Tag:   "foo",
					Count: DefaultQueryRecords,
				}
			},
		},
		{
			name:    "with filter",
			options: []QueryOption{WithFilter("foo", 42)},
			setup: func() Query {
				q := Query{Count: DefaultQueryRecords}
				q.Filters.Set("foo", 42)
				return q
			},
		},
		{
			name: "with multiple filters",
			options: []QueryOption{
				WithFilter("foo", 42),
				WithFilter("bar", "baz"),
			},
			setup: func() Query {
				q := Query{Count: DefaultQueryRecords}
				q.Filters.Set("foo", 42)
				q.Filters.Set("bar", "baz")
				return q
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			want := tc.setup()

			// Act
			q := NewQuery(tc.options...)

			// Assert
			uassert.Equal(t, want.Offset, q.Offset)
			uassert.Equal(t, want.Count, q.Count)
			uassert.Equal(t, want.Tag, q.Tag)
			uassert.Equal(t, want.Filters.Size(), q.Filters.Size())

			want.Filters.Iterate("", "", func(k string, v any) bool {
				got, exists := q.Filters.Get(k)
				uassert.True(t, exists)
				if exists {
					uassert.Equal(t, fmt.Sprint(v), fmt.Sprint(got))
				}
				return false
			})
		})
	}
}
