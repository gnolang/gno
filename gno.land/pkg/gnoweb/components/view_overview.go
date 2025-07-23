package components

import (
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

const OverviewViewType ViewType = "overview-view"

type OverviewData struct {
	PkgPath     string
	Readme      Component
	Functions   []*doc.JSONFunc
	Doc         string
	Files       []string
	FileCounter int
	Consts      []*doc.JSONValueDecl
	Vars        []*doc.JSONValueDecl
	Types       []*doc.JSONType
	Dirs        []string
}

type overviewViewParams struct {
	OverviewData
}

func OverviewView(data OverviewData) *View {
	viewData := overviewViewParams{
		OverviewData: data,
	}
	return NewTemplateView(OverviewViewType, "renderOverview", viewData)
} 