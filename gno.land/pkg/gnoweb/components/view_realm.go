package components

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
)

const RealmViewType ViewType = "realm-view"

type RealmTOCData struct {
	Items []*markdown.TocItem
}

type RealmData struct {
	Content  Component
	TocItems *RealmTOCData
}

type ArticleData struct {
	Content Component
	Classes string
}

type RealmViewData struct {
	Article ArticleData
	TOC     Component
}

func RenderRealmView(data RealmData) *View {
	viewData := RealmViewData{
		Article: ArticleData{
			Content: data.Content, // XXX:
			Classes: "realm-view lg:row-start-1",
		},
		TOC: NewTemplateComponent("renderRealmToc", data.TocItems),
	}

	return NewTemplateView(RealmViewType, "renderRealm", viewData)
}
