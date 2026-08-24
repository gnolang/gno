package query

import (
	"net/url"
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

func TestUpdateQueryValue(t *testing.T) {
	raw := "/blog?foo=bar"
	updated, err := UpdateQueryValue(raw, "foo", "baz")
	urequire.NoError(t, err, "no error")
	uassert.Equal(t, "/blog?foo=baz", updated)
}

func TestUpdateQueryFirstValues(t *testing.T) {
	raw := "/blog?sort=alpha"
	updated, err := UpdateQueryFirstValues(raw, map[string]string{
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

func TestGetQueryValueFromURL(t *testing.T) {
	raw := "/blog?view=grid&sort=alpha"
	val, err := GetQueryValueFromURL("sort", raw)
	urequire.NoError(t, err, "no error")
	uassert.Equal(t, "alpha", val)
}

func TestParseQueryFirstValuesMap(t *testing.T) {
	raw := "/blog?sort=alpha&order=desc"
	m, err := ParseQueryFirstValuesMap(raw)
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
	_, err := ParseQueryFirstValuesMap(":::")
	urequire.Error(t, err)

	_, err = UpdateQueryValue(":::", "a", "b")
	urequire.Error(t, err)

	val, err := GetQueryValueFromURL("x", ":::")
	urequire.Error(t, err)
	urequire.Equal(t, "", val)
}

func TestGetQueryValues(t *testing.T) {
	values := url.Values{
		"foo": {"1", "2"},
	}
	got := GetQueryValues(values, "foo")
	uassert.Equal(t, "1", got[0])
	uassert.Equal(t, "2", got[1])
}

func TestGetQueryFirstValue(t *testing.T) {
	values := url.Values{
		"foo": {"1", "2"},
	}
	uassert.Equal(t, "1", GetQueryFirstValue(values, "foo"))
	uassert.Equal(t, "", GetQueryFirstValue(values, "bar")) // key doesn't exist
}

func TestHasQueryKey(t *testing.T) {
	values := url.Values{
		"foo": {"1"},
	}
	uassert.Equal(t, true, HasQueryKey(values, "foo"))
	uassert.Equal(t, false, HasQueryKey(values, "bar"))
}

func TestUpdateQueryValues(t *testing.T) {
	raw := "/blog?foo=bar"
	updated, err := UpdateQueryValues(raw, "foo", []string{"1", "2"})
	urequire.NoError(t, err, "no error")
	uassert.Equal(t, "/blog?foo=1&foo=2", updated)
}

func TestUpdateQueryAllValues(t *testing.T) {
	raw := "/blog?foo=1"
	updated, err := UpdateQueryAllValues(raw, map[string][]string{
		"foo": {"2", "3"},
		"bar": {"4"},
	})
	urequire.NoError(t, err, "no error")
	// order may vary
	expected1 := "/blog?bar=4&foo=2&foo=3"
	expected2 := "/blog?foo=2&foo=3&bar=4"
	if updated != expected1 && updated != expected2 {
		t.Errorf("got %s, expected %s or %s", updated, expected1, expected2)
	}
}

func TestSetQueries(t *testing.T) {
	raw := "/blog?foo=1"
	updated, err := SetQueries(raw, map[string]string{
		"foo": "2",
		"bar": "3",
	})
	urequire.NoError(t, err, "no error")
	expected1 := "/blog?bar=3&foo=2"
	expected2 := "/blog?foo=2&bar=3"
	if updated != expected1 && updated != expected2 {
		t.Errorf("got %s, expected %s or %s", updated, expected1, expected2)
	}
}

func TestSetQueriesMulti(t *testing.T) {
	raw := "/blog?foo=1"
	updated, err := SetQueriesMulti(raw, map[string][]string{
		"foo": {"2", "3"},
		"bar": {"4"},
	})
	urequire.NoError(t, err, "no error")
	expected1 := "/blog?bar=4&foo=2&foo=3"
	expected2 := "/blog?foo=2&foo=3&bar=4"
	if updated != expected1 && updated != expected2 {
		t.Errorf("got %s, expected %s or %s", updated, expected1, expected2)
	}
}

func TestResetQuery(t *testing.T) {
	raw := "/blog?foo=1&bar=2"
	updated, err := ResetQuery(raw)
	urequire.NoError(t, err, "no error")
	uassert.Equal(t, "/blog", updated)
}
