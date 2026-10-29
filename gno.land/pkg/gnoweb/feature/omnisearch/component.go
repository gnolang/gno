package omnisearch

import "github.com/gnolang/gno/gno.land/pkg/gnoweb/components"

// NewPageView wraps the results page so callers in gnoweb's Get pipeline can
// compose it inside IndexLayout via the standard (status, *components.View)
// return shape.
func NewPageView(data SearchData) *components.View {
	return &components.View{
		Type:      OmnisearchViewType,
		Component: &components.FeatureComponent{Tmpl: pageTemplate, Name: "renderSearch", Data: data},
	}
}
