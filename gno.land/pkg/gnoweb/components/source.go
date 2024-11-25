package components

import (
	"html/template"
	"io"
)

type SourceData struct {
	PkgPath    string
	Files      []string
	FileName   string
	FileSize   float32
	FileLines  int
	FileSource template.HTML
}

func RenderSourceComponent(w io.Writer, data SourceData) error {
	return tmpl.ExecuteTemplate(w, "renderSource", data)
}
