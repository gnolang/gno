package playground

import (
	"embed"
	"html/template"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

//go:embed templates/*.html
var templateFS embed.FS

// PageTemplate is the standalone playground page. Pre-parsed at init
// so a misconfigured template surfaces immediately, not on the first
// request.
var PageTemplate = template.Must(template.New("").ParseFS(templateFS, "templates/page.html"))

// NewPageView wraps the playground template.
func NewPageView(data PlaygroundData) *components.View {
	return &components.View{
		Type:      components.PlaygroundViewType,
		Component: &components.FeatureComponent{Tmpl: PageTemplate, Name: "renderPage", Data: data},
	}
}
