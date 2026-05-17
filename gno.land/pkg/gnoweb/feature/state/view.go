package state

import (
	"html/template"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

// StateViewType re-exports the components-side constant so feature/state
// callers (component.go) don't reach across packages for the View tag.
// Both names resolve to the same underlying ViewType string.
const StateViewType = components.StateViewType

// StateData is the render payload for templates/page.html (renderPage) —
// field names must match the template.
type StateData struct {
	PkgPath    string
	Nodes      []StateNode
	CountLabel string
	Crumbs     []StateCrumb
	Sidebar    *StateSidebar

	// Height is the pinned block height (`?height=N`), 0 for latest.
	Height int64

	// LatestHref is the current URL with height stripped — "go back to live".
	LatestHref template.URL

	// ListHref is the current URL without `#fragment` — clicked to exit
	// the CSS focus mode by clearing the hash without a reload.
	ListHref template.URL

	// ViewMode is "tree" or "" (pretty default), derived from ?view= query param.
	ViewMode string

	// KindCounts feeds the kind-filter tab counters.
	KindCounts KindCounts

	// DocIndexJSON is the pre-marshaled qdoc projection over top-level
	// decls, embedded inline so the client-side controller can project
	// doc-comments onto htmx-loaded fragments without an extra RPC.
	DocIndexJSON template.JS

	// HeightParam is the resolved decimal height stamped into every
	// fragment hx-get URL so fragments inherit the parent page's
	// concrete height during nginx stale-while-revalidate windows.
	// Empty for unstamped "latest".
	HeightParam string

	// Pagination is the prev/next view-model for the top-level decls
	// footer. nil when total ≤ limit at offset 0 (no footer needed).
	Pagination *Pagination

	// SidebarTruncated is true when the full TOC exceeds maxSidebarTOC and
	// only the first cap entries are surfaced. The template renders a
	// "+N more — paginate to see them" hint when set.
	SidebarTruncated bool

	// SidebarTotal carries the realm's full top-level decl count so the
	// truncation hint can show the dropped-entry tail count (Total - cap).
	SidebarTotal int

	// SearchQuery is the validated `?search=` value. Empty when no filter
	// is active. Drives the banner + form input value in the template.
	SearchQuery string
}

// Pagination is the view-model for the top-level decls listing footer.
// Hrefs stay in the canonical `$webargs` grammar so navigation routes
// through the state handler and survives nginx caching.
type Pagination struct {
	Total       int
	StartNumber int // 1-based inclusive; may collapse to 0 on empty page
	EndNumber   int
	HasPrev     bool
	HasNext     bool
	FirstHref   template.URL
	PrevHref    template.URL
	NextHref    template.URL
	LastHref    template.URL
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
// on the corresponding row. PrettyHref/TreeHref are pre-computed by the
// sidebar builder: in-page anchors for on-page entries, cross-page
// `$state&offset=N#anchor` URLs for off-page ones — the template stamps
// them verbatim so it never has to know which kind it is rendering.
type StateTOCEntry struct {
	Label      string
	Anchor     string
	Kind       string
	Type       string
	PrettyHref template.URL
	TreeHref   template.URL
	// OnPage marks entries that resolve to an in-page row id; off-page
	// entries set data-off-page="true" on the rendered <li>.
	OnPage bool
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

// ===== htmx-fragment render payloads =====

// FragNodeData renders one node's content as a chrome-less HTML
// fragment via the shared state/nodes renderer. PkgPath + ViewMode keep
// nested permalinks correct; Height feeds sourceHref; Depth is the
// parent row's tree depth so children indent via --depth.
type FragNodeData struct {
	Node        StateNode
	PkgPath     string
	Height      int64
	HeightParam string
	ViewMode    string
	Depth       int
	// OID is the request's `?oid=…` — preserved separately because the
	// fragment's promoted root (func/closure inline) has no ObjectID of
	// its own. Drives the closure-tag OOB-swap target id.
	OID string
}

// FragSourceData feeds fragSource. SourceHTML is TRUSTED chroma markup
// — the template does not escape it. PkgPath builds the "See in code"
// permalink to the canonical full ?source view.
type FragSourceData struct {
	SourceHTML  template.HTML
	PkgPath     string
	File        string
	Line        int
	HeightParam string
}

// FragErrorData feeds fragError. Always returned with HTTP 200 so htmx
// swaps the body instead of silently dropping a 4xx/5xx.
type FragErrorData struct {
	Message   string
	RetryHint string
}
