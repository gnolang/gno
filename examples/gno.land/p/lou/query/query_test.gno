package query

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestUpdateQuery(t *testing.T) {
	raw := "/blog?foo=bar"
	updated, _ := UpdateQuery(raw, "foo", "baz")
	uassert.Equal(t, "/blog?foo=baz", updated)
}

func TestUpdateQueryMulti(t *testing.T) {
	raw := "/blog?sort=alpha"
	updated, _ := UpdateQueryMulti(raw, map[string]string{
		"sort":  "recent",
		"order": "asc",
	})
	expected1 := "/blog?order=asc&sort=recent"
	expected2 := "/blog?sort=recent&order=asc"
	if updated != expected1 && updated != expected2 {
		t.Errorf("got %s, expected %s or %s", updated, expected1, expected2)
	}
}

func TestParseQueryParams(t *testing.T) {
	raw := "/blog?view=grid&sort=alpha"
	val, _ := ParseQueryParams("sort", raw)
	uassert.Equal(t, "alpha", val)
}

func TestParseQueryMap(t *testing.T) {
	raw := "/blog?sort=alpha&order=desc"
	m, _ := ParseQueryMap(raw)
	uassert.Equal(t, "alpha", m["sort"])
	uassert.Equal(t, "desc", m["order"])
}

func TestDeleteQuery(t *testing.T) {
	raw := "/blog?sort=alpha&mode=grid"
	deleted, _ := DeleteQuery(raw, "sort")
	uassert.Equal(t, "/blog?mode=grid", deleted)
}

// func TestEmptyAndMalformedURL(t *testing.T) {
// 	uassert.Equal(t, "", ParseQueryParams("x", ":::"))
// 	uassert.Equal(t, "", UpdateQuery(":::", "a", "b"))
// }
