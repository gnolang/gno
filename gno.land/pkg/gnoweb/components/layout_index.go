package components

import (
	"context"
	"html/template"
	"io"
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

type IndexData struct {
	HeadData
	HeaderData
	FooterData
	Body template.HTML
}

func GenerateIndexData(indexData IndexData) IndexData {
	indexData.FooterData = EnrichFooterData(indexData.FooterData)
	indexData.HeaderData = EnrichHeaderData(indexData.HeaderData)
	return indexData
}

func IndexComponent(data IndexData) Component {
	return func(ctx context.Context, tmpl *template.Template, w io.Writer) error {
		return tmpl.ExecuteTemplate(w, "index", data)
	}
}

func RenderIndexComponent(w io.Writer, data IndexData) error {
	data = GenerateIndexData(data)
	return tmpl.ExecuteTemplate(w, "index", data)
}
