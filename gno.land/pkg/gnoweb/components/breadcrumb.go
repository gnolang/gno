package components

import (
	"io"
)

type BreadcrumbPart struct {
	Name string
	Path string
}

type BreadcrumbData struct {
	Parts []BreadcrumbPart
}

func RenderBreadcrumpComponent(w io.Writer, data BreadcrumbData) error {
	return tmpl.ExecuteTemplate(w, "Breadcrumb", data)
}
