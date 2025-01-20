package components

import (
	"html/template"
	"io"
)

type SourceData struct {
	PkgPath     string
	Files       []string
	FileName    string
	FileSize    string
	FileLines   int
	FileCounter int
	FileSource  template.HTML
}

func RenderSourceComponent(w io.Writer, data SourceData) error {
	return tmpl.ExecuteTemplate(w, "renderSource", data)
}
