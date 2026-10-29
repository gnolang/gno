package run

import (
	"embed"
	"html/template"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

//go:embed templates/*.html
var templateFS embed.FS

// PageTemplate is the run scratchpad page, with the shared copy button.
// Pre-parsed at init so a misconfigured template surfaces immediately, not
// on the first request.
var PageTemplate = template.Must(template.Must(
	template.New("").ParseFS(templateFS, "templates/page.html"),
).ParseFS(components.SharedPartialsFS(), "btn_copy.html"))

// NewPageView wraps the run template.
func NewPageView(data RunData) *components.View {
	return &components.View{
		Type:      components.RunViewType,
		Component: &components.FeatureComponent{Tmpl: PageTemplate, Name: "renderPage", Data: data},
	}
}
