package components

import (
	"context"
	"html/template"
	"io"

	"github.com/gnolang/gno/gno.land/pkg/markdown"
)

type RealmTOCData struct {
	Items []*markdown.TocItem
}

func RealmTOCComponent(data *RealmTOCData) Component {
	return func(ctx context.Context, tmpl *template.Template, w io.Writer) error {
		return tmpl.ExecuteTemplate(w, "renderRealmToc", data)
	}
}

func RenderRealmTOCComponent(w io.Writer, data *RealmTOCData) error {
	return tmpl.ExecuteTemplate(w, "renderRealmToc", data)
}

type RealmData struct {
	Content  template.HTML
	TocItems *RealmTOCData
}

func RenderRealmComponent(w io.Writer, data RealmData) error {
	return tmpl.ExecuteTemplate(w, "renderRealm", data)
}
