package weburl

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGnoURL_Height pins that height reads from BOTH WebQuery
// (gnoweb's `$state&height=N` syntax used by template-built URLs) and
// the standard Query (browser form GET produces `?height=N`). WebQuery
// wins when both are set since it's the gnoweb-native form.
func TestGnoURL_Height(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		web  url.Values
		std  url.Values
		want int64
	}{
		{"webquery only", url.Values{"height": {"42"}}, nil, 42},
		{"query only", nil, url.Values{"height": {"7"}}, 7},
		{"webquery wins over query", url.Values{"height": {"99"}}, url.Values{"height": {"3"}}, 99},
		{"missing", nil, nil, 0},
		{"empty value", url.Values{"height": {""}}, nil, 0},
		{"non-numeric", url.Values{"height": {"abc"}}, nil, 0},
		{"negative-looking", url.Values{"height": {"-1"}}, nil, 0},
		{"zero", url.Values{"height": {"0"}}, nil, 0},
		{"max int64-ish", url.Values{"height": {"9223372036854775807"}}, nil, 9223372036854775807},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			u := GnoURL{Path: "/r/x", WebQuery: tc.web, Query: tc.std}
			assert.Equal(t, tc.want, u.Height())
		})
	}
}

// TestGnoURL_WithHeight verifies that the WebQuery (gnoweb-native)
// form receives the height, the standard Query is cleared (avoids
// duplication), and clone semantics protect the original.
func TestGnoURL_WithHeight(t *testing.T) {
	t.Parallel()

	t.Run("sets in WebQuery, strips Query", func(t *testing.T) {
		t.Parallel()
		orig := GnoURL{
			Path:     "/r/x",
			WebQuery: url.Values{"state": {""}},
			Query:    url.Values{"height": {"3"}, "other": {"keep"}},
		}
		got := orig.WithHeight(42)
		assert.Equal(t, "42", got.WebQuery.Get("height"))
		assert.Empty(t, got.Query.Get("height"), "stale Query height stripped")
		assert.Equal(t, "keep", got.Query.Get("other"), "other Query params survive")
		assert.Empty(t, orig.WebQuery.Get("height"), "original untouched")
	})

	t.Run("h<=0 strips height", func(t *testing.T) {
		t.Parallel()
		u := GnoURL{Path: "/r/x", WebQuery: url.Values{"state": {""}, "height": {"42"}}}
		got := u.WithHeight(0)
		assert.Empty(t, got.WebQuery.Get("height"))
	})
}

// TestGnoURL_WithoutHeight verifies the "go back to live latest" URL
// strips height from BOTH WebQuery and Query while preserving every
// other parameter. Critical for object pages.
func TestGnoURL_WithoutHeight(t *testing.T) {
	t.Parallel()

	u := GnoURL{
		Path: "/r/demo/foo",
		WebQuery: url.Values{
			"state":  {""},
			"oid":    {"abc:1"},
			"tid":    {"gno.land/r/demo/foo.User"},
			"height": {"42"},
		},
		Query: url.Values{"height": {"99"}, "other": {"keep"}},
	}
	got := u.WithoutHeight()
	assert.Empty(t, got.WebQuery.Get("height"))
	assert.Empty(t, got.Query.Get("height"))
	assert.Equal(t, "abc:1", got.WebQuery.Get("oid"), "OID preserved")
	assert.Equal(t, "gno.land/r/demo/foo.User", got.WebQuery.Get("tid"), "TID preserved")
	assert.Equal(t, "keep", got.Query.Get("other"), "other Query params preserved")
	assert.Equal(t, "42", u.WebQuery.Get("height"), "original untouched (clone semantics)")
}

// TestGnoURL_Clone verifies deep-clone of the maps so mutations on
// the returned URL never alias the original.
func TestGnoURL_Clone(t *testing.T) {
	t.Parallel()

	orig := GnoURL{
		Path:     "/r/x",
		WebQuery: url.Values{"a": {"1"}},
		Query:    url.Values{"b": {"2"}},
	}
	dup := orig.Clone()
	dup.WebQuery.Set("a", "mutated")
	dup.Query.Set("b", "mutated")

	assert.Equal(t, "1", orig.WebQuery.Get("a"), "original WebQuery untouched")
	assert.Equal(t, "2", orig.Query.Get("b"), "original Query untouched")
}
