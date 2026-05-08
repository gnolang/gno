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

	// RawJSON, when set, enables the "JSON" view — a chroma-highlighted
	// pretty-printed dump of the underlying chain response. Lets devs flip
	// from the polished tree to the raw payload without leaving the page.
	// The toggle is CSS-only (no JS), driven by sibling radio inputs.
	RawJSON template.HTML

	// IsObjectPage is true when the page renders a single queried object
	// (`?state&oid=…`) rather than a realm root. The template uses it to
	// surface richer detail on the queried target — e.g. always render
	// the inline source body for a func, even though we hide it on the
	// realm root to keep the listing compact.
	IsObjectPage bool
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
