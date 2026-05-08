package components

import "html/template"

const StateViewType ViewType = "state-view"

// StateData holds everything the state explorer view needs to render the
// page server-side. It is shaped to match views/state.html field-by-field —
// the template expects exactly these names.
type StateData struct {
	// PkgPath is the realm or package path the state belongs to. Used to
	// build hrefs for stored object refs and source links.
	PkgPath string

	// Nodes is the list of top-level StateNodes (or sub-tree nodes when
	// rendering a single object's page).
	Nodes []StateNode

	// CountLabel is the human-readable summary in the header (e.g.
	// "Realm top-level declarations (5)" or "UsersMap (len=42)").
	CountLabel string

	// Crumbs is the breadcrumb path inside the state explorer. The first
	// entry is typically the realm root; subsequent entries reflect the
	// drill-down path. Empty on the realm root view.
	Crumbs []StateCrumb

	// PageNav, when set, renders previous/next links for paginated
	// collections. Nil when the collection fits in one page.
	PageNav *StatePageNav

	// Sidebar carries the side-rail content (TOC + metadata). Nil for
	// minimal views; the template skips rendering the aside when nil.
	Sidebar *StateSidebar

	// IsObjectPage is true when the page renders a single queried object
	// (`?state&oid=…`) rather than a realm root. The template uses it to
	// surface richer detail on the queried target — e.g. always render
	// the inline source body for a func, even though we hide it on the
	// realm root to keep the listing compact.
	IsObjectPage bool

	// Height is the block height the page is pinned to (`?height=N`),
	// or 0 for "latest". When non-zero, the header surfaces a chip
	// indicating time-travel mode and a link to revert to the latest.
	Height int64

	// LatestHref is the URL of the current page WITHOUT the height
	// parameter — i.e. "go back to live latest". Preserves OID/TID/
	// other query so an object-page Latest stays on that object.
	LatestHref template.URL

	// ViewMode is "tree" or "" (== pretty default), driven by the
	// `state_view_mode` cookie written by the state-view JS controller.
	// The template uses this to set the right radio's `checked`
	// attribute server-side, so the page paints in the saved view from
	// first paint — no JS-driven flicker.
	ViewMode string

	// RawJSON is the chain-native JSON response (qpkg_json or
	// qobject_json) that produced this view. Embedded as a hidden
	// element so the "Copy JSON" toolbar button can hand it to the
	// clipboard without an extra round-trip. Empty when the handler
	// chooses not to expose it (e.g., error paths).
	RawJSON string

	// KindCounts is the per-bucket count of top-level Nodes,
	// surfaced in the kind-filter tabs (`All / State / Code / Types`)
	// — same `b-tag--secondary` register as the user page's filter
	// counts. Computed by ComputeKindCounts.
	KindCounts KindCounts
}

// KindCounts holds the count of declarations falling into each
// kind bucket exposed by the state-explorer filter tabs. Buckets
// are usage-driven (not literal Kind names) so the filter UI can
// say "show me State" without enumerating every container kind.
type KindCounts struct {
	All   int // every top-level declaration
	State int // stored data: struct, map, slice, array, pointer, ref
	Code  int // executable: func, closure
	Types int // type definitions: type, interface
}

// ComputeKindCounts walks the top-level Nodes and counts each
// bucket. Reused between Pretty and Tree views; called once per
// page render in the handler.
func ComputeKindCounts(nodes []StateNode) KindCounts {
	c := KindCounts{All: len(nodes)}
	for _, n := range nodes {
		switch n.Kind {
		case "struct", "map", "slice", "array", "pointer", "ref":
			c.State++
		case "func", "closure":
			c.Code++
		case "type", "interface":
			c.Types++
		}
	}
	return c
}

// StateSidebar groups the data shown in the aside next to the state tree.
// It is built by the handler from the same StateNode list rendered in the
// main column — keeps the template free of computation.
type StateSidebar struct {
	// Heading is shown above the TOC (e.g. "Top-level declarations" on a
	// realm page, "Fields" on an object page).
	Heading string

	// TOC is the navigation list — anchor-linked entries that scroll the
	// main column to the corresponding top-level row.
	TOC []StateTOCEntry

	// Meta groups labelled facts about the page's subject (PkgPath, OID,
	// TypeID, etc.). Rendered as a compact key/value list.
	Meta []StateMetaEntry
}

// StateTOCEntry is a single entry in the side-rail navigation list. Anchor
// is matched against `id="<anchor>"` set on the corresponding row. Kind +
// Type drive the small glyph rendered next to the label so users see the
// shape (struct/map/func/etc.) at a glance.
type StateTOCEntry struct {
	Label  string
	Anchor string
	Kind   string
	Type   string
}

// StateMetaEntry is a single key/value fact in the sidebar.
type StateMetaEntry struct {
	Label string
	Value string
	// Href, when non-empty, turns the value into a link. Used to link
	// back from an object page to its realm root.
	Href template.URL
	// Mono marks values that are blockchain identifiers (OIDs, hashes,
	// etc.) — the template renders them in monospace, truncated to
	// `head…tail` with the full value reachable via title/click-to-copy.
	Mono bool
	// Section, when non-empty, opens a new visual group above this
	// entry — rendered as a small uppercase heading. Subsequent entries
	// inherit the section until the next non-empty Section.
	Section string
	// Inline renders the entry on a single line (label + value side
	// by side) instead of stacked. Useful for short values like
	// counts/sizes where a column-stack wastes vertical space.
	Inline bool
}

// StateCrumb is a single breadcrumb segment. Href is empty for the active
// (last) segment.
type StateCrumb struct {
	Label string
	Href  template.URL
}

// StatePageNav describes pagination links shown at the bottom of a paginated
// collection (large maps, slices, etc.).
type StatePageNav struct {
	PrevHref template.URL
	NextHref template.URL
	Label    string
}

// StateView creates the View for the state explorer page.
func StateView(data StateData) *View {
	return NewTemplateView(StateViewType, "renderState", data)
}
