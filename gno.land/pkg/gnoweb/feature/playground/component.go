package playground

import (
	"html/template"
	"io"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

type playgroundComponent struct {
	tmpl *template.Template
	name string
	data any
}

func (c *playgroundComponent) Render(w io.Writer) error {
	return c.tmpl.ExecuteTemplate(w, c.name, c.data)
}

// NewPageView wraps the playground template.
func NewPageView(data PlaygroundData) *components.View {
	return &components.View{
		Type: components.PlaygroundViewType,
		Component: &playgroundComponent{
			tmpl: PageTemplate,
			name: "renderPage",
			data: data,
		},
	}
}
