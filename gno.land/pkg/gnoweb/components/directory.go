package components

import (
	"io"
)

type DirData struct {
	PkgPath     string
	Files       []string
	FileCounter int
}

func RenderDirectoryComponent(w io.Writer, data DirData) error {
	return tmpl.ExecuteTemplate(w, "renderDir", data)
}
