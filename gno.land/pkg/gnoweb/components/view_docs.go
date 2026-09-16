package components

// DocsViewType identifies the docs view in the layout switch.
const DocsViewType ViewType = "docs-view"

// DocsSidebarItem is one link in the docs navigation. Href is the
// fully-resolved URL the user clicks (either /docs/<path> or an absolute
// URL for external entries); the component does no rewriting itself.
//
// Toc is non-empty only on the item the reader is currently on: the page's
// own top-level headings, nested under its entry so the rail reads as one
// "where you are" tree at two scales rather than two look-alike lists.
type DocsSidebarItem struct {
	Title    string
	Href     string
	External bool
	Active   bool
	Toc      []*TocItem
}

// DocsSidebarSection groups items under a "## ..." heading from README.md.
type DocsSidebarSection struct {
	Title string
	Items []DocsSidebarItem
}

// DocsData is the input to DocsView: rendered Markdown content plus the
// navigation tree derived from docs/README.md, with the current page's
// headings already grafted onto its entry.
type DocsData struct {
	ComponentContent Component
	Sections         []DocsSidebarSection

	// PageToc is set only when the current page appears in no section, so no
	// tree entry can carry its outline. It is then rendered as its own block.
	PageToc []*TocItem
}

type docsViewParams struct {
	Article  ArticleData
	Sections []DocsSidebarSection
	PageToc  []*TocItem
}

// DocsView returns a view that renders a documentation page with its
// navigation tree.
func DocsView(data DocsData) *View {
	p := docsViewParams{
		Article: ArticleData{
			ComponentContent: data.ComponentContent,
			Classes:          "c-realm-view",
		},
		Sections: data.Sections,
		PageToc:  data.PageToc,
	}
	return NewTemplateView(DocsViewType, "renderDocs", p)
}
