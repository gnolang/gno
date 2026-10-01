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
	// Counterpart turns the first segment (r or p) into a switch to the other
	// side of the project. Nil when the other side holds no package.
	Counterpart *HeaderLink
}

func RenderBreadcrumbComponent(w io.Writer, data BreadcrumbData) error {
	return tmpl.ExecuteTemplate(w, "Breadcrumb", data)
}
