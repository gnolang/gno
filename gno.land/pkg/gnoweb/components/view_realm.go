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
			Classes:          "c-realm-view",
		},
		ComponentTOC: NewTemplateComponent("ui/toc_realm", data.TocItems),
	}

	return NewTemplateView(RealmViewType, "renderRealm", viewData)
}
