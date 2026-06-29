package state

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// TestSliceLines pins the line-range slicer's edge cases — wrong line
// numbers from upstream must not crash or produce nonsense.
func TestSliceLines(t *testing.T) {
	t.Parallel()

	src := []byte("a\nb\nc\nd\ne\n")
	cases := []struct {
		name       string
		start, end int
		want       string
	}{
		{"normal range", 2, 4, "b\nc\nd"},
		{"single line", 3, 3, "c"},
		{"start past end-of-file", 99, 99, ""},
		{"start zero -> all", 0, 0, "a\nb\nc\nd\ne\n"},
		{"end past EOF clamps to end", 4, 999, "d\ne\n"},
		{"end < start — treat as start..eof", 4, 1, "d\ne\n"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := sliceLines(src, c.start, c.end)
			assert.Equal(t, c.want, string(got))
		})
	}
}

// TestStateObjectHref_RoundtripsThroughWebURL — the encoded ":" inside the
// OID is the load-bearing case; if URL encoding ever truncates at the
// colon the object link 404s.
func TestStateObjectHref_RoundtripsThroughWebURL(t *testing.T) {
	t.Parallel()

	const oid = "ffffffffffffffffffffffffffffffffffffffff:42"
	href := stateObjectHref("/r/demo/foo", oid, "", "")
	assert.NotEmpty(t, href)

	gnourl, err := weburl.Parse("https://gno.land" + string(href))
	require.NoError(t, err, "Href must be parsable by weburl — the very thing it routes to")
	assert.Equal(t, "/r/demo/foo", gnourl.Path)
	assert.Equal(t, oid, gnourl.WebQuery.Get("oid"),
		"ObjectID must round-trip via the URL parser without truncation at ':'")
	assert.True(t, gnourl.WebQuery.Has("state"), "state flag preserved")
}

// TestStateObjectHref_StampsTypeAndView — tid and viewMode should round-trip
// when present and stay absent when zero.
func TestStateObjectHref_StampsTypeAndView(t *testing.T) {
	t.Parallel()

	href := stateObjectHref("/r/demo", "abcd:1", "tid-x", "tree")
	gnourl, err := weburl.Parse("https://gno.land" + string(href))
	require.NoError(t, err)
	assert.Equal(t, "tid-x", gnourl.WebQuery.Get("tid"))
	assert.Equal(t, "tree", gnourl.WebQuery.Get("view"))

	bare := stateObjectHref("/r/demo", "abcd:1", "", "")
	gnourl, err = weburl.Parse("https://gno.land" + string(bare))
	require.NoError(t, err)
	assert.Empty(t, gnourl.WebQuery.Get("tid"))
	assert.Empty(t, gnourl.WebQuery.Get("view"))
}

// TestStateSourceHref_UsesWebargsGrammar — the "See in code" permalink must
// use the `$source` webargs grammar (routable) and carry the full pkg path,
// not the dead relative `?source` form (H7).
func TestStateSourceHref_UsesWebargsGrammar(t *testing.T) {
	t.Parallel()

	href := stateSourceHref("/r/demo/foo", "bar.gno", 12)
	assert.Contains(t, string(href), "/r/demo/foo$", "must use the $webargs grammar")
	assert.NotContains(t, string(href), "?source", "the dead ?query form must be gone")

	gnourl, err := weburl.Parse("https://gno.land" + string(href))
	require.NoError(t, err, "permalink must be parsable by weburl")
	assert.Equal(t, "/r/demo/foo", gnourl.Path, "full pkg path must be present, not relative")
	assert.True(t, gnourl.WebQuery.Has("source"), "source flag routes to the full-source view")
	assert.Equal(t, "bar.gno", gnourl.WebQuery.Get("file"))
	assert.True(t, strings.HasSuffix(string(href), "#L12"), "line anchor appended after encode")

	bare := stateSourceHref("/r/demo", "bar.gno", 0)
	_, err = weburl.Parse("https://gno.land" + string(bare))
	require.NoError(t, err)
	assert.NotContains(t, string(bare), "#L", "no line anchor when line is 0")
}

// TestAttachDocs_MatchByName covers the doc projection by Name against
// the union of vals + funs + typs entries. Pin the contract so a future
// refactor of the priority order (or a typo in the loop) fails here.
func TestAttachDocs_MatchByName(t *testing.T) {
	t.Parallel()

	nodes := []StateNode{
		{Name: "Counter"},
		{Name: "Render"},
		{Name: "User"},
		{Name: "Untouched"}, // no matching doc → Doc stays empty
	}

	AttachDocs(nodes,
		[]NamedDoc{{Name: "Counter", Doc: "tracks pings"}},
		[]NamedDoc{{Name: "Render", Doc: "renders the realm"}},
		[]NamedDoc{{Name: "User", Doc: "owner struct"}},
	)

	assert.Equal(t, "tracks pings", nodes[0].Doc, "val doc attached by name")
	assert.Equal(t, "renders the realm", nodes[1].Doc, "func doc attached by name")
	assert.Equal(t, "owner struct", nodes[2].Doc, "type doc attached by name")
	assert.Empty(t, nodes[3].Doc, "node with no matching doc stays empty")
}

// TestAttachDocs_EmptyDocsSkipped — empty Doc strings in the index must
// not overwrite a matching node. (The handler dedup already skips them,
// but the contract belongs on AttachDocs itself.)
func TestAttachDocs_EmptyDocsSkipped(t *testing.T) {
	t.Parallel()

	nodes := []StateNode{{Name: "X", Doc: "preexisting"}}
	AttachDocs(nodes, []NamedDoc{{Name: "X", Doc: ""}}, nil, nil)
	assert.Equal(t, "preexisting", nodes[0].Doc, "empty doc entries must not clobber")
}

// TestRecoverFetcher_LogsClippedPanic — a hostile chain returning an enormous
// panic payload must not blow up the log line itself.
func TestRecoverFetcher_LogsClippedPanic(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	func() {
		defer recoverFetcher(logger, "file", "pkgPath", "/r/x", "file", "y.gno")
		// Payload large enough to trip the 512-rune clip.
		panic(string(make([]byte, 4096)))
	}()

	out := buf.String()
	assert.Contains(t, out, "fetcher panic recovered")
	assert.Contains(t, out, "kind=file")
	// 512-rune clip + slog field/quoting overhead — line stays well under
	// the 4096-byte raw payload (would be ~4400+ if the clip didn't fire).
	assert.Less(t, len(out), 3000, "log line must not amplify the panic payload verbatim")
}

// TestRecoverToErr_NilLogger — decoder boundaries pass a nil logger;
// recoverToErr must still convert the panic into a wrapped error so the
// handler logs it through its existing error path and surfaces a clean
// 500 to the user, instead of unwinding to net/http's top-level recover
// (which logs as a process event and tears the connection).
func TestRecoverToErr_NilLogger(t *testing.T) {
	t.Parallel()

	got := func() (err error) {
		defer recoverToErr(nil, "decode pkg JSON", &err)
		panic("synthetic amino panic")
	}()

	require.Error(t, got)
	assert.Contains(t, got.Error(), "decode pkg JSON")
	assert.Contains(t, got.Error(), "panic recovered")
}

// TestRecoverToErr_NilLogger_NoPanic — happy path must leave the caller's
// named-return err untouched.
func TestRecoverToErr_NilLogger_NoPanic(t *testing.T) {
	t.Parallel()

	got := func() (err error) {
		defer recoverToErr(nil, "decode pkg JSON", &err)
		return nil
	}()

	assert.NoError(t, got)
}

// TestRecoverToErr_NilLogger_ClipsLargePayload — same 512-rune clip
// discipline as recoverFetcher: a hostile decoder panic payload cannot
// amplify the error message logged downstream.
func TestRecoverToErr_NilLogger_ClipsLargePayload(t *testing.T) {
	t.Parallel()

	got := func() (err error) {
		defer recoverToErr(nil, "decode object JSON", &err)
		panic(string(make([]byte, 4096)))
	}()

	require.Error(t, got)
	assert.Less(t, len(got.Error()), 1024, "error must not embed the full panic payload")
}

// TestComputeAnchors pins the dedup-aware fragment-anchor minting used by
// the full sidebar TOC. Mirrors buildTOC's suffix discipline so off-page
// entries land on the same row id buildTOC would have stamped.
func TestComputeAnchors(t *testing.T) {
	t.Parallel()

	got := computeAnchors([]string{"Foo", "Bar", "Foo", "Foo"})
	assert.Equal(t,
		[]string{"state-foo", "state-bar", "state-foo-1", "state-foo-2"},
		got,
	)
}

// TestFilterIndices pins the search-filter substring scan: empty query
// returns nil so the caller can fall back to the contiguous-indices path,
// matches are case-insensitive, and insertion order is preserved.
func TestFilterIndices(t *testing.T) {
	t.Parallel()

	assert.Nil(t, filterIndices([]string{"Foo", "Bar"}, ""), "empty query → nil")

	got := filterIndices([]string{"Foo", "barFoo", "baz", "FOO"}, "foo")
	assert.Equal(t, []int{0, 1, 3}, got, "case-insensitive + insertion order")

	got = filterIndices([]string{"a", "b", "c"}, "zzz")
	assert.Empty(t, got, "no-match → empty slice")
}

// TestStatePageAnchorHref locks the `#fragment` suffix grammar: anchors
// land verbatim after the encoded webargs URL, on-page page-1 URLs stay
// canonical (no offset/limit), and cross-page hops surface offset=.
func TestStatePageAnchorHref(t *testing.T) {
	t.Parallel()

	t.Run("on-page-1 pretty omits offset", func(t *testing.T) {
		t.Parallel()
		got := string(statePageAnchorHref("/r/foo", "pretty", 0, maxTopLevelDecls, "state-a-pretty"))
		assert.NotContains(t, got, "offset=")
		assert.True(t, strings.HasSuffix(got, "#state-a-pretty"))
	})

	t.Run("cross-page tree carries view+offset", func(t *testing.T) {
		t.Parallel()
		got := string(statePageAnchorHref("/r/foo", "tree", 5, maxTopLevelDecls, "state-z-tree"))
		assert.Contains(t, got, "offset=5")
		assert.Contains(t, got, "view=tree")
		assert.True(t, strings.HasSuffix(got, "#state-z-tree"))
	})
}
