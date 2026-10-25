package omnisearch

import (
	"io"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

// searchComponent renders this feature's own page template, so gnoweb's
// IndexLayout can wrap the result in the standard chrome without the
// components package needing to know the omnisearch templates.
type searchComponent struct {
	data SearchData
}

func (c *searchComponent) Render(w io.Writer) error {
	return pageTemplate.ExecuteTemplate(w, "renderSearch", c.data)
}

// NewPageView wraps the results page so callers in gnoweb's Get pipeline can
// compose it inside IndexLayout via the standard (status, *components.View)
// return shape.
func NewPageView(data SearchData) *components.View {
	return &components.View{
		Type:      OmnisearchViewType,
		Component: &searchComponent{data: data},
	}
}
