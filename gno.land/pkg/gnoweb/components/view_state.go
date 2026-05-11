package components

import "html/template"

const StateViewType ViewType = "state-view"

// StateData is the render payload for views/state.html — field names
// must match the template.
type StateData struct {
	PkgPath    string
	Nodes      []StateNode
	CountLabel string
	Crumbs     []StateCrumb
	PageNav    *StatePageNav
	Sidebar    *StateSidebar

	// IsObjectPage flips the template into single-object mode (`?state&oid=…`),
	// which surfaces richer detail than the realm root listing.
	IsObjectPage bool

	// Height is the pinned block height (`?height=N`), 0 for latest.
	Height int64

	// LatestHref is the current URL with height stripped — "go back to live".
	LatestHref template.URL

	// ViewMode is "tree" or "" (pretty default), read from the
	// state_view_mode cookie so first paint matches the saved view.
	ViewMode string

	// RawJSON is the chain-native response (qpkg_json/qobject_json),
	// embedded as a hidden element for the "Copy JSON" toolbar button.
	RawJSON string

	// KindCounts feeds the kind-filter tab counters.
	KindCounts KindCounts
}

// KindCounts counts top-level declarations per filter-tab bucket.
// Buckets are usage-driven, not literal Kind names.
type KindCounts struct {
	All   int
	State int // struct, map, slice, array, pointer, ref
	Code  int // func, closure
	Types int // type, interface
}

// ComputeKindCounts counts top-level Nodes into filter-tab buckets.
func ComputeKindCounts(nodes []StateNode) KindCounts {
	c := KindCounts{All: len(nodes)}
	for _, n := range nodes {
		switch n.Kind {
		case KindStruct, KindMap, KindSlice, KindArray, KindPointer, KindRef:
			c.State++
		case KindFunc, KindClosure:
			c.Code++
		case KindType, KindInterface:
			c.Types++
		}
	}
	return c
}

// StateSidebar is the aside content next to the state tree.
type StateSidebar struct {
	Heading string
	TOC     []StateTOCEntry
	Meta    []StateMetaEntry
}

// StateTOCEntry is a side-rail nav entry. Anchor matches `id="<anchor>"`
// on the corresponding row.
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
	// Href turns the value into a link when non-empty.
	Href template.URL
	// Mono renders blockchain IDs (OIDs, hashes) in truncated monospace.
	Mono bool
	// Section opens a new visual group; subsequent entries inherit it
	// until the next non-empty Section.
	Section string
	// Inline puts label + value on one line (compact).
	Inline bool
	// Block puts the value on its own row beneath the label (full-width).
	Block bool
}

// StateCrumb is a breadcrumb segment; Href is empty for the active segment.
type StateCrumb struct {
	Label string
	Href  template.URL
}

// StatePageNav describes pagination links for paginated collections.
type StatePageNav struct {
	PrevHref template.URL
	NextHref template.URL
	Label    string
}

// StateView creates the View for the state explorer page.
func StateView(data StateData) *View {
	return NewTemplateView(StateViewType, "renderState", data)
}
