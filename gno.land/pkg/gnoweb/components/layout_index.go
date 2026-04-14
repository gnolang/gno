package components

import "strings"

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

// BannerData holds configuration for the site-wide banner displayed above the header.
type BannerData struct {
	Text string
	URL  string
}

func (b BannerData) HasURL() bool {
	return strings.HasPrefix(b.URL, "https://") || strings.HasPrefix(b.URL, "http://")
}

func (b BannerData) Enabled() bool { return b.Text != "" }

type IndexData struct {
	HeadData
	HeaderData
	FooterData
	BodyView *View
	Mode     ViewMode
	Theme    string
	Banner   BannerData
}

type indexLayoutParams struct {
	IndexData

	// Additional data
	IsDevmodView bool
	ViewType     string
	JSController string
	Theme        string
}

func IndexLayout(data IndexData) Component {
	data.FooterData = EnrichFooterData(data.FooterData)
	data.HeaderData = EnrichHeaderData(data.HeaderData, data.Mode)

	dataLayout := indexLayoutParams{
		IndexData: data,
		ViewType:  data.BodyView.String(),
		Theme:     data.Theme,
	}

	// Set dev mode based on view type and mode
	switch data.BodyView.Type {
	case HelpViewType, SourceViewType, DirectoryViewType, StatusViewType:
		dataLayout.IsDevmodView = true
	}

	return NewTemplateComponent("index", dataLayout)
}
