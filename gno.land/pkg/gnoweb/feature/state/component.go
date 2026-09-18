package state

import (
	"html/template"
	"io"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

// stateComponent renders a template from the feature/state package's
// own template set, so the gnoweb IndexLayout can wrap the result in
// the standard chrome without the components package needing to know
// the state templates.
type stateComponent struct {
	tmpl *template.Template
	name string
	data any
}

func (c *stateComponent) Render(w io.Writer) error {
	return c.tmpl.ExecuteTemplate(w, c.name, c.data)
}

// NewPageView wraps the full-page state template so callers in
// gnoweb's Get pipeline can compose it inside IndexLayout via the
// standard (status, *components.View) return shape.
func NewPageView(data StateData) *components.View {
	return &components.View{
		Type:      StateViewType,
		Component: &stateComponent{tmpl: PageTemplate, name: "renderPage", data: data},
	}
}
