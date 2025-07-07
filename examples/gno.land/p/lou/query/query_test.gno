package query

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestUpdateQuery(t *testing.T) {
	raw := "/blog?foo=bar"
	updated := updateQuery(raw, "foo", "baz")
	uassert.Equal(t, "/blog?foo=baz", updated)
}

func TestUpdateQueryMulti(t *testing.T) {
	raw := "/blog?sort=alpha"
	updated := updateQueryMulti(raw, map[string]string{
		"sort":  "recent",
		"order": "asc",
	})
	uassert.Equal(t, "/blog?order=asc&sort=recent", updated)
}
