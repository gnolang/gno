package state

import "github.com/gnolang/gno/gno.land/pkg/gnoweb/components"

// NewPageView wraps the full-page state template so callers in
// gnoweb's Get pipeline can compose it inside IndexLayout via the
// standard (status, *components.View) return shape.
func NewPageView(data StateData) *components.View {
	return &components.View{
		Type:             StateViewType,
		SkipTargetInBody: true, // on the content header
		Component:        &components.FeatureComponent{Tmpl: PageTemplate, Name: "renderPage", Data: data},
	}
}
