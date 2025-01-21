package components

import (
	"bytes"
	"html/template"
	"io"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
)

type RealmTOCData struct {
	Items []*markdown.TocItem
}

type RealmData struct {
	Content  template.HTML
	TocItems *RealmTOCData
}

type ArticleData struct {
	Content template.HTML
	Classes string
}

type RealmViewData struct {
	Article ArticleData
	TOC     template.HTML
}

func RenderRealmComponent(w io.Writer, data RealmData) error {
	var tocBuf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&tocBuf, "renderRealmToc", data.TocItems); err != nil {
		return err
	}

	viewData := RealmViewData{
		Article: ArticleData{
			Content: data.Content,
			Classes: "realm-content lg:row-start-1",
		},
		TOC: template.HTML(tocBuf.String()), //nolint:gosec
	}

	return tmpl.ExecuteTemplate(w, "renderRealm", viewData)
}
