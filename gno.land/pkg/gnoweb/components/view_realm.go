package components

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
)

const RealmViewType ViewType = "realm-view"

type RealmTOCData struct {
	Items []*markdown.TocItem
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
			Classes:          "realm-view lg:row-start-1 pt-6 lg:pt-10",
		},
		ComponentTOC: NewTemplateComponent("ui/toc_realm", data.TocItems),
	}

	return NewTemplateView(RealmViewType, "renderRealm", viewData)
}
