package components

// Layout
const (
	SidebarLayout = "sidebar"
	FullLayout    = "full"
)

// ViewMode represents the current view mode of the application
// It affects the layout, navigation, and display of content
type ViewMode int

const (
	ViewModeExplorer ViewMode = iota // For exploring packages and paths
	ViewModeRealm                    // For realm content display
	ViewModePackage                  // For package content display
	ViewModeHome                     // For home page display
)

// View mode predicates
func (m ViewMode) IsExplorer() bool { return m == ViewModeExplorer }
func (m ViewMode) IsRealm() bool    { return m == ViewModeRealm }
func (m ViewMode) IsPackage() bool  { return m == ViewModePackage }
func (m ViewMode) IsHome() bool     { return m == ViewModeHome }

// GetLayoutType returns the appropriate layout type for the view mode
func (m ViewMode) GetLayoutType() string {
	switch m {
	case ViewModeRealm, ViewModeHome, ViewModePackage:
		return SidebarLayout
	default:
		return FullLayout
	}
}

// ShouldShowDevTools returns whether dev tools should be shown for this mode
func (m ViewMode) ShouldShowDevTools() bool {
	return m != ViewModeHome
}

// ShouldShowGeneralLinks returns whether general navigation links should be shown
func (m ViewMode) ShouldShowGeneralLinks() bool {
	return m == ViewModeHome
}

type HeadData struct {
	Title       string
	Description string
	Canonical   string
	Image       string
	URL         string
	ChromaPath  string
	AssetsPath  string
	Analytics   bool
	Remote      string
	ChainId     string
}

type IndexData struct {
	HeadData
	HeaderData
	FooterData
	BodyView *View
	Mode     ViewMode
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
	data.HeaderData = EnrichHeaderData(data.HeaderData, data.Mode)

	dataLayout := indexLayoutParams{
		IndexData: data,
		ViewType:  data.BodyView.String(),
	}

	// Set layout based on view type and mode
	if data.BodyView.Type == DirectoryViewType || data.Mode == ViewModeExplorer {
		dataLayout.Layout = FullLayout
	} else {
		dataLayout.Layout = data.Mode.GetLayoutType()
	}

	// Set dev mode based on view type and mode
	switch data.BodyView.Type {
	case HelpViewType, SourceViewType, DirectoryViewType, StatusViewType:
		dataLayout.IsDevmodView = true
	}

	return NewTemplateComponent("index", dataLayout)
}
