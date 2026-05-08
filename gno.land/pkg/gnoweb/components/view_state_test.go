package components

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// renderState executes the renderState template against the given data and
// returns the produced HTML. Test-only helper.
func renderState(t *testing.T, data StateData) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, tmpl.ExecuteTemplate(&buf, "renderState", data))
	return buf.String()
}

// TestRenderState_Empty covers the case of a realm with no exposed state —
// the template should emit the "no exposed state" placeholder, not an empty
// tree container that would mislead users into thinking the page is broken.
func TestRenderState_Empty(t *testing.T) {
	t.Parallel()

	html := renderState(t, StateData{
		PkgPath:    "/r/demo/empty",
		Nodes:      nil,
		CountLabel: "Realm top-level declarations (0)",
	})

	assert.Contains(t, html, "no exposed state",
		"empty packages should render an explanatory placeholder")
	assert.NotContains(t, html, `class="tree"`,
		"don't emit an empty tree container when there's nothing to show")
}

// TestRenderState_LeafPrimitive verifies that a single primitive top-level
// decl renders as a `.b-state-decl` card (matching the action view pattern)
// — not as a tree row. The card header carries name + type, the content
// holds the value as a `<code class="value">` block.
func TestRenderState_LeafPrimitive(t *testing.T) {
	t.Parallel()

	html := renderState(t, StateData{
		PkgPath: "/r/demo/foo",
		Nodes: []StateNode{
			{Name: "x", Type: "int", Kind: "primitive", Value: "42"},
		},
		CountLabel: "Realm top-level declarations (1)",
	})

	assert.Contains(t, html, `<article class="b-state-decl"`,
		"top-level decl wraps in a card, not a tree row")
	assert.Contains(t, html, ">x</h2>",
		"variable name surfaces as the card title")
	assert.Contains(t, html, `data-kind="primitive"`)
	assert.Contains(t, html, `class="value"`)
	assert.Contains(t, html, ">42<")
	// No <details> inside the explorer body for primitives. The page
	// header has a `b-state-history` <details> for time-travel —
	// that's unrelated.
	bodyStart := strings.Index(html, `<article class="b-state-explorer">`)
	if bodyStart >= 0 {
		body := html[bodyStart:]
		assert.NotContains(t, body, "<details", "a primitive isn't expandable")
	}
}

// TestRenderState_RefAsLink verifies that a stored object reference is
// rendered as a navigable <a> with the correct href — that's the core of
// the navigation-driven UX (replaces the previous JS-driven AJAX expand).
func TestRenderState_RefAsLink(t *testing.T) {
	t.Parallel()

	// Hrefs are built by the orchestrator (Enrich), not the walker — the
	// template just renders what it's given. Mimic the orchestrator's output
	// here so the test stays scoped to template behaviour.
	const oid = "ffffffffffffffffffffffffffffffffffffffff:42"
	nodes := []StateNode{{
		Name: "Users", Type: "map[string]User", Kind: "ref",
		ObjectID: oid, Expandable: true,
	}}
	Enrich(nodes, "/r/demo/foo", 0, nil, nil) // nil deps: link build only

	html := renderState(t, StateData{
		PkgPath:    "/r/demo/foo",
		Nodes:      nodes,
		CountLabel: "Realm top-level declarations (1)",
	})

	assert.Contains(t, html, `<article class="b-state-decl"`)
	assert.Contains(t, html, ">Users</h2>", "name surfaces as card title")
	assert.Contains(t, html, `>Open<`,
		"card header carries a labeled Open button to the dedicated page")
	assert.Contains(t, html, "oid=ffffffffffffffffffffffffffffffffffffffff%3A42",
		"href URL-encodes ':' so the parser doesn't truncate the OID")
	// A bare stored ref with no inline children must not emit a <details>
	// inside the explorer body. The page header carries an unrelated
	// `<details class="b-state-history">` for time-travel — that's
	// fine, it's not part of the state-decl rendering.
	bodyStart := strings.Index(html, `<article class="b-state-explorer">`)
	if bodyStart >= 0 {
		body := html[bodyStart:]
		assert.NotContains(t, body, "<details", "no <details> inside the explorer body for a bare ref")
	}
}

// TestRenderState_BranchExpandable verifies that a top-level branch
// renders as a card whose body holds the children tree (one level of
// nested branches use <details>/<summary> for CSS-only toggle).
func TestRenderState_BranchExpandable(t *testing.T) {
	t.Parallel()

	html := renderState(t, StateData{
		PkgPath: "/r/demo/foo",
		Nodes: []StateNode{{
			Name: "Config", Type: "Config", Kind: "struct",
			Expandable: true,
			Children: []StateNode{
				{Name: "MaxItems", Type: "int", Kind: "primitive", Value: "100"},
				{Name: "Owner", Type: "string", Kind: "primitive", Value: `"alice"`},
				{Name: "Inner", Type: "Inner", Kind: "struct", Expandable: true,
					Children: []StateNode{{Name: "x", Type: "int", Kind: "primitive", Value: "1"}}},
			},
		}},
		CountLabel: "Realm top-level declarations (1)",
	})

	assert.Contains(t, html, `<article class="b-state-decl"`,
		"top-level branch renders as a card")
	assert.Contains(t, html, ">Config</h2>")
	// Direct children are flat rows inside the card body.
	assert.Contains(t, html, "MaxItems")
	assert.Contains(t, html, "100")
	assert.Contains(t, html, "alice")
	// Nested branches still get <details>/<summary>.
	assert.Contains(t, html, "<details")
	assert.Contains(t, html, `data-shape="branch"`)
	// Depth in the children tree starts at 0 inside the card.
	assert.Contains(t, html, "--depth: 0;")
}

// TestRenderState_ClosureCaptures verifies the closure-specific UI: the
// "Captured variables:" label appears when kind=closure and there are
// children, helping developers identify what state a closure carries.
func TestRenderState_ClosureCaptures(t *testing.T) {
	t.Parallel()

	html := renderState(t, StateData{
		PkgPath: "/r/demo/foo",
		Nodes: []StateNode{{
			Name: "stepper", Type: "func() int", Kind: "closure",
			Expandable: true,
			Children: []StateNode{
				{Name: "value", Type: "int", Kind: "primitive", Value: "5"},
			},
		}},
		CountLabel: "Realm top-level declarations (1)",
	})

	assert.Contains(t, html, "Captured variables")
	assert.Contains(t, html, "stepper")
	assert.Contains(t, html, ">5<")
}

// TestRenderState_HTMLEscaping is the load-bearing security test: any user-
// controlled value (Name, Value, Type, ObjectID) flowing into the template
// must be HTML-escaped automatically by html/template. A regression here
// would mean a malicious realm could inject scripts into gnoweb's origin.
func TestRenderState_HTMLEscaping(t *testing.T) {
	t.Parallel()

	html := renderState(t, StateData{
		PkgPath: "/r/demo/foo",
		Nodes: []StateNode{
			{Name: `<script>alert(1)</script>`, Type: "string", Kind: "primitive",
				Value: `"<img src=x onerror=alert(1)>"`},
		},
		CountLabel: "Realm top-level declarations (1)",
	})

	// Raw payload bytes must NOT appear unescaped.
	assert.NotContains(t, html, "<script>alert(1)</script>",
		"script tags from realm content must be escaped, not rendered as HTML")
	assert.NotContains(t, html, "<img src=x onerror=alert(1)>",
		"img injection from realm value must be escaped")
	// Properly escaped versions should appear.
	assert.Contains(t, html, "&lt;script&gt;",
		"name field is escaped via html/template")
	assert.Contains(t, html, "&lt;img",
		"value field is escaped via html/template")
}

// TestRenderState_PageNav verifies that pagination renders prev/next links
// when set, and skips the section entirely when nil.
func TestRenderState_PageNav(t *testing.T) {
	t.Parallel()

	html := renderState(t, StateData{
		PkgPath:    "/r/demo/foo",
		Nodes:      []StateNode{{Name: "k1", Type: "int", Kind: "primitive", Value: "1"}},
		CountLabel: "Map (len=200, page 2/4)",
		PageNav: &StatePageNav{
			PrevHref: "/r/demo/foo$state&oid=ffffffffffffffffffffffffffffffffffffffff:8&page=1",
			NextHref: "/r/demo/foo$state&oid=ffffffffffffffffffffffffffffffffffffffff:8&page=3",
			Label:    "page 2 of 4",
		},
	})

	assert.Contains(t, html, "previous")
	assert.Contains(t, html, "next")
	assert.Contains(t, html, "page 2 of 4")
	assert.Contains(t, html, `rel="prev"`)
	assert.Contains(t, html, `rel="next"`)
}

// TestRenderState_PageNav_None covers the unpaginated case — the nav should
// not appear at all.
func TestRenderState_PageNav_None(t *testing.T) {
	t.Parallel()

	html := renderState(t, StateData{
		PkgPath:    "/r/demo/foo",
		Nodes:      []StateNode{{Name: "x", Type: "int", Kind: "primitive", Value: "1"}},
		CountLabel: "Realm top-level declarations (1)",
		PageNav:    nil,
	})

	assert.NotContains(t, html, "previous")
	assert.NotContains(t, html, "next")
	assert.NotContains(t, html, `class="pagenav"`)
}

// TestRenderState_Crumbs verifies that breadcrumbs render only when set,
// with the last (active) segment as plain text and earlier segments as
// links — preserving navigation context when drilling into objects.
func TestRenderState_Crumbs(t *testing.T) {
	t.Parallel()

	html := renderState(t, StateData{
		PkgPath: "/r/demo/foo",
		Nodes:   []StateNode{},
		Crumbs: []StateCrumb{
			{Label: "/r/demo/foo", Href: "/r/demo/foo$state"},
			{Label: "Users"},
		},
		CountLabel: "UsersMap (len=42)",
	})

	assert.Contains(t, html, "/r/demo/foo")
	assert.Contains(t, html, "Users")
	assert.Contains(t, html, `href="/r/demo/foo$state"`,
		"first crumb is a link back to the realm root")
	// The last crumb has no Href and should not be wrapped in <a>.
	idx := strings.LastIndex(html, "Users")
	require.GreaterOrEqual(t, idx, 0)
	tail := html[idx:]
	openTag := strings.LastIndex(html[:idx], "<a")
	closeTag := strings.LastIndex(html[:idx], "</a>")
	assert.GreaterOrEqual(t, closeTag, openTag,
		"the last crumb 'Users' must not be inside an open <a> tag")
	_ = tail
}

// TestStateView_Factory verifies the public factory wires the data through
// to a renderable View. Production code paths in handler_http.go go through
// here, so a sanity check beats relying on integration tests alone.
func TestStateView_Factory(t *testing.T) {
	t.Parallel()

	view := StateView(StateData{
		PkgPath:    "/r/demo/foo",
		Nodes:      []StateNode{{Name: "x", Type: "int", Kind: "primitive", Value: "1"}},
		CountLabel: "Realm top-level declarations (1)",
	})
	require.NotNil(t, view)

	var buf bytes.Buffer
	require.NoError(t, view.Render(&buf))

	body := buf.String()
	assert.Contains(t, body, "b-state-explorer")
	assert.Contains(t, body, ">x<")
	assert.Contains(t, body, ">1<")
}

// TestRenderState_Sidebar_TOCAndAnchors verifies that when a Sidebar is
// supplied, the aside renders with TOC links pointing to anchor ids stamped
// on the matching top-level rows. Critical for keyboard / no-JS users to
// jump between declarations on a large realm page.
func TestRenderState_Sidebar_TOCAndAnchors(t *testing.T) {
	t.Parallel()

	nodes := []StateNode{
		{Name: "Counter", Type: "int", Kind: "primitive", Value: "42"},
		{Name: "Users", Type: "map[string]User", Kind: "ref",
			ObjectID: "ffffffffffffffffffffffffffffffffffffffff:8", Expandable: true},
	}
	sidebar := BuildPackageSidebar("/r/demo/foo", nodes)
	require.NotNil(t, sidebar)

	html := renderState(t, StateData{
		PkgPath:    "/r/demo/foo",
		Nodes:      nodes,
		CountLabel: "Realm top-level declarations (2)",
		Sidebar:    sidebar,
	})

	assert.Contains(t, html, `<aside class="b-sidebar sidebar"`,
		"sidebar uses gnoweb's existing .b-sidebar block, not a state-specific class")
	assert.Contains(t, html, "Top-level declarations", "sidebar heading")
	// TOC entries emit one anchor per view (Pretty + Tree); CSS hides
	// the variant for the inactive view. Each variant points to a
	// distinct id so we never have duplicate ids in the DOM.
	assert.Contains(t, html, `href="#state-counter-pretty"`)
	assert.Contains(t, html, `href="#state-counter-tree"`)
	assert.Contains(t, html, `id="state-counter-pretty"`)
	assert.Contains(t, html, `id="state-counter-tree"`)
	assert.Contains(t, html, `href="#state-users-pretty"`)
	assert.Contains(t, html, `href="#state-users-tree"`)
	assert.Contains(t, html, `id="state-users-pretty"`)
	assert.Contains(t, html, `id="state-users-tree"`)
	// Meta surfaces the realm path.
	assert.Contains(t, html, "Realm")
	assert.Contains(t, html, "/r/demo/foo")
}

// TestRenderState_Sidebar_AnchorsOnlyOnTopLevel ensures nested children
// don't accidentally collide with top-level anchor ids — the template
// stamps `id` only when the recursion is at the root.
func TestRenderState_Sidebar_AnchorsOnlyOnTopLevel(t *testing.T) {
	t.Parallel()

	nodes := []StateNode{{
		Name: "Outer", Type: "struct", Kind: "struct", Expandable: true,
		Children: []StateNode{
			{Name: "Inner", Type: "int", Kind: "primitive", Value: "1"},
		},
	}}
	sidebar := BuildPackageSidebar("/r/demo/foo", nodes)

	html := renderState(t, StateData{
		PkgPath: "/r/demo/foo", Nodes: nodes,
		CountLabel: "Realm top-level declarations (1)", Sidebar: sidebar,
	})

	// Top-level rows get per-view ids (`-pretty` + `-tree`); nested
	// children must not be stamped with any anchor id.
	assert.Contains(t, html, `id="state-outer-pretty"`)
	assert.Contains(t, html, `id="state-outer-tree"`)
	assert.NotContains(t, html, `id="state-inner"`,
		"deep children must not be stamped with anchor ids")
}

// TestRenderState_Sidebar_ObjectMetaLinksBack verifies the per-object
// sidebar carries a link back to the realm root state page — preserves
// navigation context after drilling into an object.
func TestRenderState_Sidebar_ObjectMetaLinksBack(t *testing.T) {
	t.Parallel()

	nodes := []StateNode{
		{Name: "Name", Type: "string", Kind: "primitive", Value: `"alice"`},
	}
	sidebar := BuildObjectSidebar(
		"/r/demo/foo",
		"ffffffffffffffffffffffffffffffffffffffff:1",
		"gno.land/r/demo/foo.User",
		0, // height = 0 (latest)
		StateObjectInfoView{},
		nodes,
	)

	html := renderState(t, StateData{
		PkgPath: "/r/demo/foo", Nodes: nodes,
		CountLabel: "Object", Crumbs: nil, Sidebar: sidebar,
	})

	assert.Contains(t, html, "Object ID")
	assert.Contains(t, html, "ffffffffffffffffffffffffffffffffffffffff:1")
	assert.Contains(t, html, "Type")
	assert.Contains(t, html, "gno.land/r/demo/foo.User")
	// Back link to the realm.
	assert.Contains(t, html, `href="/r/demo/foo$state"`)
}

// TestRenderState_OpenDepthsByDefault locks the per-view UX contract:
//   - Pretty view opens depth 0 + 1 (cards expect their content open
//     so the user immediately sees the structure).
//   - Tree view opens depth 0 only (the flat tree should stay tight,
//     letting the user pick what to drill into rather than presenting
//     a half-expanded wall).
//
// Both views render to the same DOM; CSS toggles visibility, so the
// count is summed across both.
func TestRenderState_OpenDepthsByDefault(t *testing.T) {
	t.Parallel()

	html := renderState(t, StateData{
		PkgPath: "/r/demo/foo",
		Nodes: []StateNode{{
			Name: "Top", Type: "struct", Kind: "struct", Expandable: true,
			Children: []StateNode{{
				Name: "L1", Type: "struct", Kind: "struct", Expandable: true,
				Children: []StateNode{{
					Name: "L2", Type: "struct", Kind: "struct", Expandable: true,
					Children: []StateNode{{
						Name: "L3", Type: "struct", Kind: "struct", Expandable: true,
						Children: []StateNode{
							{Name: "leaf", Type: "int", Kind: "primitive", Value: "1"},
						},
					}},
				}},
			}},
		}},
		CountLabel: "Realm top-level declarations (1)",
	})

	// Scope to the explorer body — the page header has an unrelated
	// `<details class="b-state-history">` for time-travel that we
	// don't want to count here.
	bodyStart := strings.Index(html, `<article class="b-state-explorer">`)
	require.GreaterOrEqual(t, bodyStart, 0, "explorer article must be present")
	body := html[bodyStart:]

	// Robust counting via regex — `open` may appear after other attrs
	// (e.g. `<details open class="nested">` for Pretty fields).
	reOpen := regexp.MustCompile(`<details\s[^>]*\bopen\b[^>]*>|<details\s+open(?:\s|>)`)
	reAny := regexp.MustCompile(`<details(?:\s|>)`)
	openCount := len(reOpen.FindAllString(body, -1))
	totalCount := len(reAny.FindAllString(body, -1))
	closedCount := totalCount - openCount
	// Pretty view (card content starts at depth 0): L1=open, L2=open,
	//   L3=closed. (3 details, 2 open — depth < 2)
	// Tree view (state/nodes from the root): Top=open, L1=closed, L2=closed,
	//   L3=closed. (4 details, 1 open — depth < 1)
	// Both rendered → 3 open, 4 closed.
	assert.Equal(t, 3, openCount, "Pretty=depth<2 + Tree=depth<1 details open across both views")
	assert.Equal(t, 4, closedCount, "remaining details closed across both views")
}

// TestRenderState_StatsStrip locks the audit metadata rendering as a
// labeled stat-block grid (.b-state-stats) on the card itself: each
// fact gets a small uppercase title (`<dt class="title">LABEL</dt>`)
// above its value (`<dd class="value">data</dd>`), matching the
// action-card PARAMS/COMMAND typography. Absent fields are skipped
// so the strip stays compact.
func TestRenderState_StatsStrip(t *testing.T) {
	t.Parallel()

	const oid = "ffffffffffffffffffffffffffffffffffffffff:42"
	nodes := []StateNode{{
		Name: "config", Type: "*Config", Kind: "ref",
		ObjectID: oid, Expandable: true,
		Hash:           "8f3e",
		OwnerID:        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:1",
		RefCount:       "3",
		LastObjectSize: "412",
		ModTime:        "14237",
	}}
	Enrich(nodes, "/r/demo/foo", 0, nil, nil)

	html := renderState(t, StateData{
		PkgPath: "/r/demo/foo", Nodes: nodes,
		CountLabel:   "Object 1 fields",
		IsObjectPage: true, // chips only render on the dedicated object page
	})

	assert.Contains(t, html, `class="b-state-stats"`,
		"audit metadata renders as a stat-block grid on the object page")
	assert.Contains(t, html, ">OID</dt>", "OID stat title present")
	assert.Contains(t, html, ">Size</dt>", "Size stat title present")
	assert.Contains(t, html, "412 B", "size formatted with B suffix")
	assert.Contains(t, html, ">Refs</dt>", "Refs stat title present")
	assert.Contains(t, html, ">3</dd>", "RefCount value present")
	assert.Contains(t, html, ">Owner</dt>", "Owner stat title present")
	assert.Contains(t, html, ">Hash</dt>", "Hash stat title present")
	assert.Contains(t, html, ">8f3e<", "hash value present")
	assert.Contains(t, html, ">Modified</dt>", "Modified stat title present")
	assert.Contains(t, html, ">#14237<", "ModTime prefixed with # (block height)")
	// Owner stat links into the owner's own state page (URL-encoded
	// via weburl.GnoURL; `:` becomes `%3A`, keys sorted alphabetically).
	assert.Contains(t, html,
		`href="/r/demo/foo$oid=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa%3A1&amp;state"`,
		"Owner stat is a link to the owner's state page")
}

// TestRenderState_StatsStrip_OmittedForPrimitives pins the auto-skip
// behaviour: nodes with no audit fields (primitives like int, bool,
// string, plus type declarations) MUST NOT emit a stat-strip — those
// don't have OIDs/Hashes/etc, and an empty strip would be visual noise.
// Stored objects (slices, maps, structs, refs) carry the chips.
func TestRenderState_StatsStrip_OmittedForPrimitives(t *testing.T) {
	t.Parallel()

	html := renderState(t, StateData{
		PkgPath: "/r/demo/foo",
		Nodes: []StateNode{
			{Name: "Counter", Type: "int", Kind: "primitive", Value: "42"},
			{Name: "Active", Type: "bool", Kind: "primitive", Value: "true"},
		},
		CountLabel: "Realm top-level declarations (2)",
	})

	assert.NotContains(t, html, `class="b-state-stats"`,
		"primitives have no audit metadata — strip omitted")
}

// TestRenderState_StatsStrip_Empty verifies that when no audit fields are
// set, the strip is not rendered at all — the card stays minimal for
// inline-only values.
func TestRenderState_StatsStrip_Empty(t *testing.T) {
	t.Parallel()

	html := renderState(t, StateData{
		PkgPath: "/r/demo/foo",
		Nodes: []StateNode{
			{Name: "x", Type: "int", Kind: "primitive", Value: "42"},
		},
		CountLabel: "Realm top-level declarations (1)",
	})

	assert.NotContains(t, html, `class="b-state-stats"`,
		"primitive cards must not emit an empty stats strip")
}

// TestRenderState_FooterCTA verifies the compact footer CTA — the
// escape hatch that pushes users into the dedicated object page.
// Only rendered on cards that have an Href (stored refs); absent
// otherwise so primitive cards stay clean.
func TestRenderState_FooterCTA(t *testing.T) {
	t.Parallel()

	const oid = "ffffffffffffffffffffffffffffffffffffffff:42"
	nodes := []StateNode{{
		Name: "config", Type: "*Config", Kind: "ref",
		ObjectID: oid, Expandable: true,
	}}
	Enrich(nodes, "/r/demo/foo", 0, nil, nil)

	html := renderState(t, StateData{
		PkgPath: "/r/demo/foo", Nodes: nodes,
		CountLabel: "Realm top-level declarations (1)",
	})

	assert.Contains(t, html, `class="cta"`,
		"stored refs emit a compact footer CTA")
	assert.Contains(t, html, "Inspect ref",
		"CTA label includes the kind so the verb reads naturally")
}

// TestRenderState_CTA_PlusNMore exercises the "+N more" badge when the
// inline preview shows fewer entries than Length advertises. Pins the
// `derefInt` template helper — the bug was: `with .Length` on a `*int`
// passed the POINTER (not the dereffed int) to `gt`, raising "invalid
// type for comparison" at execute time and aborting the entire render.
func TestRenderState_CTA_PlusNMore(t *testing.T) {
	t.Parallel()

	totalLen := 12
	nodes := []StateNode{{
		Name: "Catalog", Type: "*avl.Tree", Kind: "ref",
		ObjectID: "ffffffffffffffffffffffffffffffffffffffff:7",
		Expandable: true,
		Length:     &totalLen,
		Children: []StateNode{
			{Name: "0", Type: "string", Kind: "primitive", Value: "first"},
			{Name: "1", Type: "string", Kind: "primitive", Value: "second"},
			{Name: "2", Type: "string", Kind: "primitive", Value: "third"},
		},
	}}
	Enrich(nodes, "/r/demo/foo", 0, nil, nil)

	html := renderState(t, StateData{
		PkgPath: "/r/demo/foo", Nodes: nodes,
		CountLabel: "Realm top-level declarations (1)",
	})

	assert.Contains(t, html, `class="more"`, "+N more badge present when truncated")
	assert.Contains(t, html, "+9 more", "shows count of hidden entries (12 total - 3 shown)")
	assert.Contains(t, html, "Inspect ref", "main CTA label still rendered")
}

// TestRenderState_CTA_NoMore verifies the "+N more" badge is OMITTED
// when all children fit (Length == len(Children)) — keeps the CTA
// minimal in the common case.
func TestRenderState_CTA_NoMore(t *testing.T) {
	t.Parallel()

	totalLen := 2
	nodes := []StateNode{{
		Name: "Pair", Type: "[2]int", Kind: "ref",
		ObjectID: "ffffffffffffffffffffffffffffffffffffffff:1",
		Expandable: true,
		Length:     &totalLen,
		Children: []StateNode{
			{Name: "0", Type: "int", Kind: "primitive", Value: "1"},
			{Name: "1", Type: "int", Kind: "primitive", Value: "2"},
		},
	}}
	Enrich(nodes, "/r/demo/foo", 0, nil, nil)

	html := renderState(t, StateData{
		PkgPath: "/r/demo/foo", Nodes: nodes,
		CountLabel: "Realm top-level declarations (1)",
	})

	assert.NotContains(t, html, `class="more"`, "no +N more when all children shown")
}

// TestRenderState_NoFooterCTAForPrimitive verifies a value-only card has
// no footer CTA — there's nothing to inspect on a primitive int.
func TestRenderState_NoFooterCTAForPrimitive(t *testing.T) {
	t.Parallel()

	html := renderState(t, StateData{
		PkgPath: "/r/demo/foo",
		Nodes: []StateNode{
			{Name: "x", Type: "int", Kind: "primitive", Value: "42"},
		},
		CountLabel: "Realm top-level declarations (1)",
	})

	assert.NotContains(t, html, `class="cta"`,
		"primitive cards must not emit a CTA banner — nowhere to navigate")
}

// TestRenderState_PreviewInRow verifies the walker-built Preview field
// surfaces as `<span class="preview">` next to the type in collapsed
// rows — "config : Config {name: …}" instead of just "config : Config".
func TestRenderState_PreviewInRow(t *testing.T) {
	t.Parallel()

	html := renderState(t, StateData{
		PkgPath: "/r/demo/foo",
		Nodes: []StateNode{{
			Name: "u", Type: "User", Kind: "struct",
			Expandable: true,
			Preview:    `{Name: "alice", Age: 30}`,
			Children: []StateNode{
				{Name: "Name", Type: "string", Kind: "primitive", Value: `"alice"`},
				{Name: "Age", Type: "int", Kind: "primitive", Value: "30"},
			},
		}},
		CountLabel: "Realm top-level declarations (1)",
	})

	assert.Contains(t, html, `class="preview"`,
		"Preview field renders inside row summary")
	assert.Contains(t, html, `Name: &#34;alice&#34;`,
		"preview content escaped through html/template")
}

// TestRenderState_PreviewWalker verifies the walker auto-builds Preview
// strings for struct nodes — no test fixture should need to set Preview
// manually except in template-only tests.
func TestRenderState_PreviewWalker(t *testing.T) {
	t.Parallel()

	// Decoded children: 4 fields, only first 3 surface in preview.
	children := []StateNode{
		{Name: "name", Type: "string", Kind: "primitive", Value: `"alice"`},
		{Name: "age", Type: "int", Kind: "primitive", Value: "30"},
		{Name: "ok", Type: "bool", Kind: "primitive", Value: "true"},
		{Name: "extra", Type: "string", Kind: "primitive", Value: `"hidden"`},
	}
	preview := buildChildrenPreview(children)
	assert.Contains(t, preview, `name: "alice"`)
	assert.Contains(t, preview, `age: 30`)
	assert.Contains(t, preview, `ok: true`)
	assert.NotContains(t, preview, `hidden`,
		"preview truncates beyond inlinePreviewMaxFields")
	assert.Contains(t, preview, "…", "ellipsis marks truncation")
}

// TestRenderState_NoBTagOnTypes pins the editorial-restraint design
// choice: no .b-tag pill chrome on the type label inside the
// explorer article. The type reads as italic mono text inline.
// Filter tabs in the sub-header use `.b-tag--secondary` for kind
// counts (mirrors user-page filters); we scope the assertion to
// the article content only.
func TestRenderState_NoBTagOnTypes(t *testing.T) {
	t.Parallel()

	html := renderState(t, StateData{
		PkgPath: "/r/demo/foo",
		Nodes: []StateNode{{
			Name: "Outer", Type: "struct", Kind: "struct", Expandable: true,
			Children: []StateNode{
				{Name: "inner", Type: "int", Kind: "primitive", Value: "1"},
			},
		}},
		CountLabel: "Realm top-level declarations (1)",
	})

	// Scope to the explorer article — kind-filter pills live in the
	// sub-header which is allowed to use `b-tag--secondary`.
	articleStart := strings.Index(html, `<article class="b-state-explorer">`)
	require.GreaterOrEqual(t, articleStart, 0, "explorer article must be present")
	article := html[articleStart:]
	assert.NotContains(t, article, `b-tag`,
		"the editorial-restraint design strips .b-tag pill chrome from types")
}

// TestRenderState_NoUnexpectedControllers locks the JS surface to the
// short list of allowed controllers. Adding a controller is fine but
// must be intentional — adding one without updating this list catches
// it in code review.
func TestRenderState_NoUnexpectedControllers(t *testing.T) {
	t.Parallel()

	html := renderState(t, StateData{
		PkgPath: "/r/demo/foo",
		Nodes: []StateNode{
			{Name: "x", Type: "int", Kind: "primitive", Value: "42"},
			{Name: "Users", Type: "map[string]User", Kind: "ref",
				ObjectID: "ffffffffffffffffffffffffffffffffffffffff:8", Expandable: true},
		},
		CountLabel: "Realm top-level declarations (2)",
	})

	allowed := map[string]bool{
		`data-controller="copy"`:                 true, // OID/value/hash click-to-copy
		`data-controller="state-view"`:           true, // Pretty/Tree cookie+localStorage persistence
		`data-controller="state-search"`:         true, // in-page filter by decl name
		`data-controller="state-tree"`:           true, // persist tree expand/collapse per OID + sidebar TOC scroll bridge
		`data-controller="state-tree-controls"`: true, // expand-all / collapse-all toolbar
	}
	for _, line := range strings.Split(html, "\n") {
		if !strings.Contains(line, "data-controller=") {
			continue
		}
		ok := false
		for needle := range allowed {
			if strings.Contains(line, needle) {
				ok = true
				break
			}
		}
		assert.True(t, ok, "unexpected data-controller in line: %q", strings.TrimSpace(line))
	}
}
