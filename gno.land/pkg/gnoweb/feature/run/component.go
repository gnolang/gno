package run

import (
	"html/template"
	"io"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

type runComponent struct {
	tmpl *template.Template
	name string
	data any
}

func (c *runComponent) Render(w io.Writer) error {
	return c.tmpl.ExecuteTemplate(w, c.name, c.data)
}

// NewPageView wraps the run template.
func NewPageView(data RunData) *components.View {
	return &components.View{
		Type: components.RunViewType,
		Component: &runComponent{
			tmpl: PageTemplate,
			name: "renderPage",
			data: data,
		},
	}
}
