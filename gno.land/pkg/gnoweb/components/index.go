package components

import (
	"context"
	"html/template"
	"io"
	"net/url"
)

type HeadData struct {
	Title       string
	Description string
	Canonical   string
	Image       string
	URL         string
	ChromaPath  string
	AssetsPath  string
	Analytics   bool
}

type HeaderData struct {
	RealmPath  string
	Breadcrumb BreadcrumbData
	WebQuery   url.Values
}

type FooterData struct {
	Analytics  bool
	AssetsPath string
}

type IndexData struct {
	HeadData
	HeaderData
	FooterData
	Body template.HTML
}

func IndexComponent(data IndexData) Component {
	return func(ctx context.Context, tmpl *template.Template, w io.Writer) error {
		return tmpl.ExecuteTemplate(w, "index", data)
	}
}

func RenderIndexComponent(w io.Writer, data IndexData) error {
	return tmpl.ExecuteTemplate(w, "index", data)
}
