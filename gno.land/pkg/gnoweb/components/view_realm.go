package components

const RealmViewType ViewType = "realm-view"

// TocItem represents a table of contents item for the components package.
type TocItem struct {
	Title string
	ID    string
	Items []*TocItem
}

// Anchor returns the anchor link for this ToC item.
func (i TocItem) Anchor() string {
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
