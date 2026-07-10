package components

const RealmViewType ViewType = "realm-view"

// TocItem represents a table of contents item for the components package.
type TocItem struct {
	Title string
	ID    string
	// Icon is an optional sprite id suffix (e.g. "kind-func") rendered as a
	// leading kind glyph. Empty means no glyph — realm/action/source TOCs leave
	// it unset, so only the package overview surfaces icons.
	Icon string
	// Link points off the current page, e.g. at a file in the source view. When
	// it is empty the item anchors to its own ID.
	Link  string
	Items []*TocItem
}

// Anchor returns this ToC item's target: its Link when set, otherwise an anchor
// to its own ID.
func (i TocItem) Anchor() string {
	if i.Link != "" {
		return i.Link
	}
	return "#" + i.ID
}

type RealmTOCData struct {
	Items []*TocItem
}

type RealmData struct {
	ComponentContent Component
	TocItems         *RealmTOCData
}

type ArticleData struct {
	ComponentContent Component
	Classes          string
}

type realmViewParams struct {
	Article      ArticleData
	ComponentTOC Component
}

func RealmView(data RealmData) *View {
	viewData := realmViewParams{
		Article: ArticleData{
			ComponentContent: data.ComponentContent,
			Classes:          "c-realm-view",
		},
		ComponentTOC: NewTemplateComponent("ui/toc_realm", data.TocItems),
	}

	return NewTemplateView(RealmViewType, "renderRealm", viewData)
}
