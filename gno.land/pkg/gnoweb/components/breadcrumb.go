package components

import (
	"io"
)

type BreadcrumbPart struct {
	Name string
	URL  string
}

type BreadcrumbData struct {
	Parts []BreadcrumbPart
	Args  string
}

func RenderBreadcrumpComponent(w io.Writer, data BreadcrumbData) error {
	return tmpl.ExecuteTemplate(w, "Breadcrumb", data)
}
