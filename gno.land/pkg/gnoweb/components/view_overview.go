package components

import (
	"io"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

const OverviewViewType ViewType = "overview-view"

// DocRenderer renders markdown doc strings and source code to HTML.
// RenderSource applies syntax highlighting based on the file extension in name.
// Implementations must be safe for concurrent use and HTML-safe by construction.
type DocRenderer interface {
	RenderDocumentation(w io.Writer, src []byte) error
	RenderSource(w io.Writer, name string, src []byte) error
}

// License describes a detected license file.
// Kind is empty when the file exists but its license type is unknown.
type License struct {
	Kind     string
	FileName string
}

// PackageInfo carries identity metadata displayed in the sidebar.
type PackageInfo struct {
	Namespace   string
	PackagePath string
	PackageType string // "realm" | "pure"
	License     License
	GnoVersion  string
	Creator     string // gnomod [addpkg] creator address (on-chain deploys)
	Height      int    // gnomod [addpkg] deploy block height
	Draft       bool   // gnomod draft = not production-ready
	Private     bool   // gnomod private
}

// PackageStats aggregates numeric counters derived from files and qdoc.
type PackageStats struct {
	FileCount     int
	GnoFileCount  int
	TestCount     int
	FuncCount     int
	ExportedFunc  int
	TypeCount     int
	ConstCount    int
	VarCount      int
	ImportCount   int
	CrossingCount int
}

// PackageQuality exposes boolean presence flags used to render ✓/✗ indicators.
type PackageQuality struct {
	HasReadme  bool
	HasTests   bool
	HasLicense bool
	HasPkgDoc  bool
}

// FuncEntry is the view-owned representation of a function or method.
type FuncEntry struct {
	Name               string
	SignatureComponent Component
	Doc                Component
	Receiver           string
	Crossing           bool
	IsMethod           bool
	ActionURL          string
	AnchorID           string
	SourceURL          string // links to the exact file + line in source view
}

// TypeEntry is the view-owned representation of a type declaration.
type TypeEntry struct {
	Name               string
	SignatureComponent Component
	Doc                Component
	Kind               string
	Methods            []FuncEntry
	AnchorID           string
	SourceURL          string
}

// ValueGroup is a const/var declaration group preserving source order.
type ValueGroup struct {
	Kind               string // "const" | "var"
	Names              string
	SignatureComponent Component
	Doc                Component
	AnchorID           string
	SourceURL          string
}

// ImportLink is a dependency edge rendered in the Imports section.
type ImportLink struct {
	Path     string
	Kind     string // "stdlib" | "package" | "realm" | "external"
	Link     string
	External bool
}

// PathSegments is Path cut after each slash, for a template to offer a line
// break there before any inside a segment.
func (l ImportLink) PathSegments() []string { return PathSegments(l.Path) }

// PathSegments cuts a package path after each slash. A long path then wraps
// between segments, so an address stays whole on its line where it fits:
// two lookalike addresses are compared at a glance only when unbroken.
func PathSegments(p string) []string { return strings.SplitAfter(p, "/") }

// FileLink is a file entry rendered in the Files section.
type FileLink struct {
	Name      string
	Link      string
	IsTest    bool
	IsReadme  bool
	IsLicense bool
}

// SubpackageLink is a direct child package rendered in the Subdirectories section.
type SubpackageLink struct {
	Name     string
	Path     string
	Synopsis string
}

// OverviewInput aggregates the data required to build an OverviewData.
type OverviewInput struct {
	URL         *weburl.GnoURL
	Files       []string
	Doc         *doc.JSONDocumentation
	Sources     map[string][]byte
	Subpaths    []string
	Readme      Component
	Domain      string
	DocRenderer DocRenderer
	// DepsURL is the page listing the package's importers, set only when
	// this deployment can compute them.
	DepsURL string
	// Storage is the realm's stored size and deposit; nil when the node did
	// not say.
	Storage *PackageStorage
	// Calls is the Recent calls section; nil without an indexer.
	Calls *CallsSection
}

// OverviewData is the full payload passed to the overview template.
type OverviewData struct {
	PkgPath    string
	Title      string
	Synopsis   string
	PackageDoc Component
	Readme     Component

	Info    PackageInfo
	Stats   PackageStats
	Quality PackageQuality

	Funcs       []FuncEntry
	Types       []TypeEntry
	Consts      []ValueGroup
	Vars        []ValueGroup
	Files       []FileLink
	Subpackages []SubpackageLink
	Bugs        []string

	// SymbolsTruncated is set when funcs/types/values were capped at
	// maxOverviewSymbols; the template then shows a "view full source" notice.
	SymbolsTruncated bool

	// Graph draws the package between what it imports and, when an indexer
	// can tell, what imports it.
	Graph DepGraph

	// Storage and Calls are passed through from OverviewInput.
	Storage *PackageStorage
	Calls   *CallsSection

	ComponentTOC Component
}

// OverviewView constructs a new overview View from pre-built data.
func OverviewView(data OverviewData) *View {
	view := NewTemplateView(OverviewViewType, "renderOverview", data)
	view.SkipTargetInBody = true // on the content header
	return view
}
