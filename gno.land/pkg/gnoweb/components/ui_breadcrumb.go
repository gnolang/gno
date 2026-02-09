package components

import (
	"io"
)

type BreadcrumbPart struct {
	Name string
	URL  string
}

type QueryParam struct {
	Key   string
	Value string
}

type BreadcrumbData struct {
	Parts    []BreadcrumbPart
	ArgParts []BreadcrumbPart
	Queries  []QueryParam
}

func RenderBreadcrumbComponent(w io.Writer, data BreadcrumbData) error {
	return tmpl.ExecuteTemplate(w, "Breadcrumb", data)
}
