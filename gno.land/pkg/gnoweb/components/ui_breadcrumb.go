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
	// Namespace is set on /r/<ns>/... and /p/<ns>/... paths, where the first
	// segment opens a menu: the listing it used to link, the namespace page,
	// and the Counterpart when there is one.
	Namespace string
	// Counterpart heads that menu with the other side of the project. Nil
	// when the other side holds no package.
	Counterpart *HeaderLink
}

func RenderBreadcrumbComponent(w io.Writer, data BreadcrumbData) error {
	return tmpl.ExecuteTemplate(w, "Breadcrumb", data)
}
