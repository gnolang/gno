package components

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
	BodyView *View
}

type indexLayoutParams struct {
	IndexData

	// Additional data
	IsDevmodView bool
	Layout       string
	ViewType     string
}

func IndexLayout(data IndexData) Component {
	data.FooterData = EnrichFooterData(data.FooterData)
	data.HeaderData = EnrichHeaderData(data.HeaderData)

	dataLayout := indexLayoutParams{
		IndexData: data,
		// Set default value
		Layout:   FullLayout,
		ViewType: data.BodyView.String(),
	}

	switch data.BodyView.Type {
	case RealmViewType:
		dataLayout.Layout = SidebarLayout

	case HelpViewType:
		dataLayout.IsDevmodView = true
		dataLayout.Layout = SidebarLayout

	case SourceViewType:
		dataLayout.IsDevmodView = true
		dataLayout.Layout = SidebarLayout

	case DirectoryViewType:
		dataLayout.IsDevmodView = true

	case StatusViewType:
		dataLayout.IsDevmodView = true
	}

	return NewTemplateComponent("index", dataLayout)
}
