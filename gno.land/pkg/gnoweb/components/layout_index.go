package components

import (
	"html/template"
	"io"
)

// Layout
const (
	SidebarLayout = "sidebar"
	FullLayout    = "full"
)

type HeadData struct {
	Title       string
	Description string
	Canonical   string
	Image       string
	URL         string
	ChromaPath  string
	AssetsPath  string
	Analytics   bool
}

type IndexData struct {
	HeadData
	HeaderData
	FooterData
	*View
	// Body          template.HTML
}

func registerIndexFuncs(funcs template.FuncMap) {
	funcs["IndexGetLayout"] = func(i IndexData) string {
		switch i.View.Type {
		case RealmViewType, HelpViewType, SourceViewType, DirectoryViewType:
			return SidebarLayout
		default:
			return FullLayout
		}
	}

	funcs["IndexIsDevmodView"] = func(i IndexData) bool {
		switch i.View.Type {
		case SourceViewType, HelpViewType, DirectoryViewType, StatusViewType:
			return true
		default:
			return false
		}
	}
}

func GenerateIndexData(indexData IndexData) IndexData {
	return indexData
}

func IndexLayout(data IndexData) Component {
	data.FooterData = EnrichFooterData(data.FooterData)
	data.HeaderData = EnrichHeaderData(data.HeaderData)

	return NewTemplateComponent("index", data)
}

func RenderIndexLayout(w io.Writer, data IndexData) error {
	return IndexLayout(data).Render(w)
}
