package components

// ViewMode represents the current view mode of the application
// It affects the layout, navigation, and display of content
type ViewMode int

const (
	ViewModeExplorer ViewMode = iota // For exploring packages and paths
	ViewModeRealm                    // For realm content display
	ViewModePackage                  // For package content display
	ViewModeHome                     // For home page display
	ViewModeUser                     // For user page display
)

// View mode predicates
func (m ViewMode) IsExplorer() bool { return m == ViewModeExplorer }
func (m ViewMode) IsRealm() bool    { return m == ViewModeRealm }
func (m ViewMode) IsPackage() bool  { return m == ViewModePackage }
func (m ViewMode) IsUser() bool     { return m == ViewModeUser }
func (m ViewMode) IsHome() bool     { return m == ViewModeHome }

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
	BuildTime   string
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
	ViewType     string
	JSController string
}

func IndexLayout(data IndexData) Component {
	data.FooterData = EnrichFooterData(data.FooterData)
	data.HeaderData = EnrichHeaderData(data.HeaderData, data.Mode)

	dataLayout := indexLayoutParams{
		IndexData: data,
		ViewType:  data.BodyView.String(),
	}

	// Set dev mode based on view type and mode
	switch data.BodyView.Type {
	case HelpViewType, SourceViewType, DirectoryViewType, StatusViewType:
		dataLayout.IsDevmodView = true
	}

	return NewTemplateComponent("index", dataLayout)
}
