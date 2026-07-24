package components

// DocsViewType identifies the docs view in the layout switch.
const DocsViewType ViewType = "docs-view"

// DocsSidebarItem is one link in the docs navigation. Href is the
// fully-resolved URL the user clicks (either /docs/<path> or an absolute
// URL for external entries); the component does no rewriting itself.
type DocsSidebarItem struct {
	Title    string
	Href     string
	External bool
	Active   bool
}

// DocsSidebarSection groups items under a "## ..." heading from README.md.
type DocsSidebarSection struct {
	Title string
	Items []DocsSidebarItem
}

// DocsData is the input to DocsView: rendered Markdown content plus the
// section navigation derived from docs/README.md. The right-side in-page
// TOC is intentionally dropped here in favor of a single sidebar showing
// the section navigation; doc pages are typically short and consistent
// cross-page navigation matters more than per-page outline.
type DocsData struct {
	ComponentContent Component
	Sections         []DocsSidebarSection
}

type docsViewParams struct {
	Article  ArticleData
	Sections []DocsSidebarSection
}

// DocsView returns a view that renders a documentation page with the
// section sidebar derived from README.md.
func DocsView(data DocsData) *View {
	p := docsViewParams{
		Article: ArticleData{
			ComponentContent: data.ComponentContent,
			Classes:          "c-realm-view",
		},
		Sections: data.Sections,
	}
	return NewTemplateView(DocsViewType, "renderDocs", p)
}
