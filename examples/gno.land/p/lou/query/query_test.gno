package query

import (
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

func TestUpdateQuery(t *testing.T) {
	raw := "/blog?foo=bar"
	updated, err := UpdateQuery(raw, "foo", "baz")
	urequire.NoError(t, err, "no error")
	uassert.Equal(t, "/blog?foo=baz", updated)
}

func TestUpdateQueryMulti(t *testing.T) {
	raw := "/blog?sort=alpha"
	updated, err := UpdateQueryMulti(raw, map[string]string{
		"sort":  "recent",
		"order": "asc",
	})
	urequire.NoError(t, err, "no error")
	expected1 := "/blog?order=asc&sort=recent"
	expected2 := "/blog?sort=recent&order=asc"
	if updated != expected1 && updated != expected2 {
		t.Errorf("got %s, expected %s or %s", updated, expected1, expected2)
	}
}

func TestParseQueryParams(t *testing.T) {
	raw := "/blog?view=grid&sort=alpha"
	val, err := ParseQueryParams("sort", raw)
	urequire.NoError(t, err, "no error")
	uassert.Equal(t, "alpha", val)
}

func TestParseQueryMap(t *testing.T) {
	raw := "/blog?sort=alpha&order=desc"
	m, err := ParseQueryMap(raw)
	urequire.NoError(t, err, "no error")
	uassert.Equal(t, "alpha", m["sort"])
	uassert.Equal(t, "desc", m["order"])
}

func TestDeleteQuery(t *testing.T) {
	raw := "/blog?sort=alpha&mode=grid"
	deleted, err := DeleteQuery(raw, "sort")
	urequire.NoError(t, err, "no error")
	uassert.Equal(t, "/blog?mode=grid", deleted)
}

func TestMalformedURLHandling(t *testing.T) {
	_, err := ParseQueryMap(":::")
	urequire.Error(t, err)

	_, err = UpdateQuery(":::", "a", "b")
	urequire.Error(t, err)

	val, err := ParseQueryParams("x", ":::")
	urequire.Error(t, err)
	urequire.Equal(t, "", val)
}
