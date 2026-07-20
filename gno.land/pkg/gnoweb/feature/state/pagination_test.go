package state

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildManyTopLevelDeclsFixture writes a qpkg_json with `n` top-level int
// decls "v0".."v(n-1)" so pagination tests can scan the slice boundaries.
// Mirrors buildDeepStructFixture's amino-JSON shape so the same decoder
// path is exercised.
func buildManyTopLevelDeclsFixture(n int) []byte {
	var b strings.Builder
	b.WriteString(`{"names":[`)
	for i := range n {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, `"v%d"`, i)
	}
	b.WriteString(`],"values":[`)
	for i := range n {
		if i > 0 {
			b.WriteString(",")
		}
		nbuf := make([]byte, 8)
		for j := range 8 {
			nbuf[j] = byte(i >> (8 * j))
		}
		fmt.Fprintf(&b,
			`{"T":{"@type":"/gno.PrimitiveType","value":"32"},"N":"%s"}`,
			base64.StdEncoding.EncodeToString(nbuf),
		)
	}
	b.WriteString(`]}`)
	return []byte(b.String())
}

// TestDecodePackagePagination scans the slice boundaries on a known
// fixture so a regression in paginationWindow / DecodePackage is caught
// by exact-name assertions, not just length checks.
func TestDecodePackagePagination(t *testing.T) {
	t.Parallel()

	fixture := buildManyTopLevelDeclsFixture(12)

	t.Run("first page", func(t *testing.T) {
		t.Parallel()
		nodes, total, err := DecodePackage(context.Background(), fixture, DefaultPageRenderConfig(), 0, 5)
		require.NoError(t, err)
		assert.Equal(t, 12, total)
		require.Len(t, nodes, 5)
		assert.Equal(t, "v0", nodes[0].Name)
		assert.Equal(t, "v4", nodes[4].Name)
	})

	t.Run("middle page", func(t *testing.T) {
		t.Parallel()
		nodes, total, err := DecodePackage(context.Background(), fixture, DefaultPageRenderConfig(), 5, 5)
		require.NoError(t, err)
		assert.Equal(t, 12, total)
		require.Len(t, nodes, 5)
		assert.Equal(t, "v5", nodes[0].Name)
		assert.Equal(t, "v9", nodes[4].Name)
	})

	t.Run("last partial page", func(t *testing.T) {
		t.Parallel()
		nodes, total, err := DecodePackage(context.Background(), fixture, DefaultPageRenderConfig(), 10, 5)
		require.NoError(t, err)
		assert.Equal(t, 12, total)
		require.Len(t, nodes, 2)
		assert.Equal(t, "v10", nodes[0].Name)
		assert.Equal(t, "v11", nodes[1].Name)
	})

	t.Run("out-of-range offset", func(t *testing.T) {
		t.Parallel()
		nodes, total, err := DecodePackage(context.Background(), fixture, DefaultPageRenderConfig(), 99, 5)
		require.NoError(t, err)
		assert.Equal(t, 12, total, "total stays honest even with out-of-range offset")
		assert.Empty(t, nodes)
	})

	t.Run("negative offset clamps", func(t *testing.T) {
		t.Parallel()
		nodes, _, err := DecodePackage(context.Background(), fixture, DefaultPageRenderConfig(), -1, 5)
		require.NoError(t, err)
		require.Len(t, nodes, 5)
		assert.Equal(t, "v0", nodes[0].Name)
	})

	t.Run("limit zero defaults to cap", func(t *testing.T) {
		t.Parallel()
		nodes, _, err := DecodePackage(context.Background(), fixture, DefaultPageRenderConfig(), 0, 0)
		require.NoError(t, err)
		assert.Len(t, nodes, min(maxTopLevelDecls, 12))
	})
}

// TestBuildPaginationHrefs locks the prev/next view-model: bounds,
// HasPrev/HasNext, and the href grammar (canonical $webargs with offset
// omitted on page 0, view= preserved). Also confirms First/Last hrefs
// are NOT built on page 1 / Prev/Next on the last page — the template
// never renders them, so allocating them is wasted work.
func TestBuildPaginationHrefs(t *testing.T) {
	t.Parallel()

	t.Run("first page of three", func(t *testing.T) {
		t.Parallel()
		p := buildPagination("/r/foo", "pretty", "", 12, 0, 5)
		require.NotNil(t, p)
		assert.Equal(t, 12, p.Total)
		assert.Equal(t, 1, p.StartNumber)
		assert.Equal(t, 5, p.EndNumber)
		assert.False(t, p.HasPrev)
		assert.True(t, p.HasNext)
		// Hrefs only built for sides the template will render.
		assert.Empty(t, p.FirstHref, "no FirstHref on page 1 (template gates on HasPrev)")
		assert.Empty(t, p.PrevHref, "no PrevHref on page 1")
		assert.Contains(t, string(p.NextHref), "offset=5")
		assert.Contains(t, string(p.LastHref), "offset=10")
	})

	t.Run("middle page", func(t *testing.T) {
		t.Parallel()
		p := buildPagination("/r/foo", "pretty", "", 12, 5, 5)
		require.NotNil(t, p)
		assert.True(t, p.HasPrev)
		assert.True(t, p.HasNext)
		// Prev is 0 — encoded as no offset param (canonical first-page URL).
		assert.NotContains(t, string(p.PrevHref), "offset=")
		assert.Contains(t, string(p.NextHref), "offset=10")
	})

	t.Run("last page", func(t *testing.T) {
		t.Parallel()
		p := buildPagination("/r/foo", "pretty", "", 12, 10, 5)
		require.NotNil(t, p)
		assert.Equal(t, 11, p.StartNumber)
		assert.Equal(t, 12, p.EndNumber)
		assert.True(t, p.HasPrev)
		assert.False(t, p.HasNext)
		assert.Empty(t, p.NextHref, "no NextHref on the last page")
		assert.Empty(t, p.LastHref, "no LastHref on the last page")
	})

	t.Run("no pagination when total ≤ limit at offset 0", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, buildPagination("/r/foo", "pretty", "", 3, 0, 5))
	})

	t.Run("view preserved on every href", func(t *testing.T) {
		t.Parallel()
		p := buildPagination("/r/foo", "tree", "", 12, 0, 5)
		require.NotNil(t, p)
		assert.Contains(t, string(p.NextHref), "view=tree")
		assert.Contains(t, string(p.LastHref), "view=tree")
	})

	t.Run("out-of-range offset still renders honest 0-0 summary", func(t *testing.T) {
		t.Parallel()
		p := buildPagination("/r/foo", "pretty", "", 12, 99, 5)
		require.NotNil(t, p)
		assert.Equal(t, 0, p.StartNumber, "no rows shown → start collapses to 0")
		assert.Equal(t, 12, p.EndNumber, "end clamped to total")
	})

	t.Run("active search is threaded into every page href", func(t *testing.T) {
		t.Parallel()
		// Every page href must carry search= so a filtered list stays filtered.
		p := buildPagination("/r/foo", "pretty", "needle", 12, 5, 5)
		require.NotNil(t, p)
		assert.Contains(t, string(p.FirstHref), "search=needle")
		assert.Contains(t, string(p.PrevHref), "search=needle")
		assert.Contains(t, string(p.NextHref), "search=needle")
		assert.Contains(t, string(p.LastHref), "search=needle")
	})
}

// TestServePackageSearchPaginationKeepsFilter renders the package page under
// a search spanning multiple pages and asserts the pagination footer keeps
// search= so Next/Last stay filtered.
func TestServePackageSearchPaginationKeepsFilter(t *testing.T) {
	t.Parallel()

	// 12 decls v0..v11 all match search "v"; limit 5 paginates the matches.
	client := &pageMockClient{pkgBytes: buildManyTopLevelDeclsFixture(12)}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{"search": {"v"}, "limit": {"5"}}, "/r/demo")
	require.Equal(t, 200, rec.Code, "body=%s", rec.Body.String())

	body := rec.Body.String()
	assert.Contains(t, body, "offset=5", "a Next page must render under search")
	assert.Contains(t, body, "search=v",
		"pagination hrefs must keep the active search (only page hrefs carry search= in the page body)")
}

// TestStatePageHrefOmitsDefaults confirms the canonical-URL discipline:
// page-1 / default-limit URLs match the unparameterized state URL so
// nginx cache keys stay parity. Diverging params (custom limit, view
// mode) MUST surface so they don't drop on page hops.
func TestStatePageHrefOmitsDefaults(t *testing.T) {
	t.Parallel()

	// Page 1, pretty → canonical (no offset, no limit, no view).
	href := string(statePageHref("/r/foo", "pretty", "", 0, maxTopLevelDecls))
	assert.NotContains(t, href, "offset=")
	assert.NotContains(t, href, "limit=")
	assert.NotContains(t, href, "view=")

	// Non-default everything → all params surface.
	href = string(statePageHref("/r/foo", "tree", "search", 5, 3))
	assert.Contains(t, href, "search=search")
	assert.Contains(t, href, "offset=5")
	assert.Contains(t, href, "limit=3")
	assert.Contains(t, href, "view=tree")
}

func TestLastPageOffset(t *testing.T) {
	t.Parallel()

	cases := []struct {
		total, limit, want int
	}{
		{0, 5, 0},
		{1, 5, 0},
		{5, 5, 0},
		{6, 5, 5},
		{12, 5, 10},
		{15, 5, 10},
		{16, 5, 15},
		{100, 0, 0}, // limit ≤ 0 short-circuits to 0
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, lastPageOffset(tc.total, tc.limit),
			"lastPageOffset(total=%d, limit=%d)", tc.total, tc.limit)
	}
}

// TestServeJSONPackagePaginated drives the paginated json.go path through
// the shared mock-client helper from json_test.go and asserts the wrapper
// shape + slice boundaries on a 12-decl fixture. Single source of truth
// for handler test plumbing — no parallel stub hierarchy.
func TestServeJSONPackagePaginated(t *testing.T) {
	t.Parallel()

	h := newJSONHandler(&mockJSONClient{pkgBytes: buildManyTopLevelDeclsFixture(12)})
	rec := serveJSONReq(t, h, url.Values{"offset": {"5"}, "limit": {"3"}})
	require.Equal(t, 200, rec.Code, "body=%s", rec.Body.String())

	var got pkgJSONWrapper
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "/r/demo", got.PkgPath)
	assert.Equal(t, 12, got.Total)
	assert.Equal(t, 5, got.Offset)
	assert.Equal(t, 3, got.Limit)
	require.Len(t, got.Names, 3)
	assert.Equal(t, []string{"v5", "v6", "v7"}, got.Names)
	require.Len(t, got.Values, 3, "Values slice mirrors Names slice")
}

// TestServeJSONPackageInvalidOffset400 / InvalidLimit400 — bad pagination
// input must 400, never silently degrade to default-paged page 1.
func TestServeJSONPackageInvalidOffset400(t *testing.T) {
	t.Parallel()

	h := newJSONHandler(&mockJSONClient{pkgBytes: buildManyTopLevelDeclsFixture(12)})
	rec := serveJSONReq(t, h, url.Values{"offset": {"-1"}})
	assert.Equal(t, 400, rec.Code)
}

func TestServeJSONPackageInvalidLimit400(t *testing.T) {
	t.Parallel()

	h := newJSONHandler(&mockJSONClient{pkgBytes: buildManyTopLevelDeclsFixture(12)})
	rec := serveJSONReq(t, h, url.Values{"limit": {"abc"}})
	assert.Equal(t, 400, rec.Code)
}

// TestTemplateStampsViewModeLiteralPerContainer pins the fix for the
// "tree mode shows pretty fragments" bug. The Pretty/Tree CSS toggle is
// client-side: switching it does NOT re-render the page or rewrite the
// pre-stamped hx-get URLs. If the SSR stamped each container's lazy
// fragments with the page's URL-derived ViewMode, a user loading
// ?view=pretty then toggling to Tree would see pretty markup hydrated
// inside the tree container (and vice versa). Each container must stamp
// its own literal view mode regardless of the page's ?view= param.
func TestTemplateStampsViewModeLiteralPerContainer(t *testing.T) {
	t.Parallel()

	// Fixture: one ref-typed top-level decl so both containers emit
	// stateFragNodeHref-stamped <details>.
	const refOnly = `{
		"names": ["myRef"],
		"values": [{"T":{"@type":"/gno.RefType","ID":"x.T"},"V":{"@type":"/gno.RefValue","ObjectID":"715383ba05505afed61caa873216e2ee896bede9:10"}}]
	}`

	cases := []struct {
		name    string
		urlView string
	}{
		{"page loaded as pretty", ""},
		{"page loaded as tree", "tree"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := &pageMockClient{pkgBytes: []byte(refOnly)}
			h := newPageHandler(client)
			q := url.Values{}
			if tc.urlView != "" {
				q.Set("view", tc.urlView)
			}
			rec := servePageReq(t, h, q, "/r/demo")
			require.Equal(t, 200, rec.Code, "body=%s", rec.Body.String())
			body := rec.Body.String()

			// Container affinity: the view-pretty <div> always stamps
			// pretty fragments (no view=tree), the view-tree <div>
			// always stamps tree fragments (with view=tree). Regardless
			// of the URL viewMode.
			prettyChunk := sliceBetween(body, `<div class="view-pretty">`, `<div class="view-tree"`)
			treeChunk := sliceBetween(body, `<div class="view-tree"`, `</article>`)

			assert.NotContains(t, prettyChunk, "view=tree",
				"view-pretty container must never stamp view=tree on hx-get")
			assert.Contains(t, treeChunk, "view=tree",
				"view-tree container must always stamp view=tree on hx-get")
		})
	}
}

// sliceBetween returns the substring between (exclusive of) the first
// occurrence of `start` and the next occurrence of `end`. Returns empty
// if either marker is missing. Used by the view-mode test to scope grep
// to each container's markup.
func sliceBetween(s, start, end string) string {
	_, after, ok := strings.Cut(s, start)
	if !ok {
		return ""
	}
	rest := after
	before, _, ok := strings.Cut(rest, end)
	if !ok {
		return rest
	}
	return before
}
