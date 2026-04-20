package components

// AnalyticsContext is the audience-context label on AnalyticsData.
type AnalyticsContext string

const (
	AnalyticsContextBuilder AnalyticsContext = "builder"
	AnalyticsContextNeutral AnalyticsContext = "neutral"
)

// AnalyticsData holds the SimpleAnalytics metadata rendered into the page.
type AnalyticsData struct {
	Enabled    bool
	PageType   string
	Context    AnalyticsContext
	ChainId    string
	AssetsPath string
	BuildTime  string
}

// ClassifyView returns the page-type label and AnalyticsContext for a given
// mode and view. View type takes precedence when both match: a Source view
// inside a Realm mode is "source", not "realm".
func ClassifyView(mode ViewMode, view ViewType) (string, AnalyticsContext) {
	switch view {
	case SourceViewType:
		return "source", AnalyticsContextBuilder
	case HelpViewType:
		return "help", AnalyticsContextBuilder
	case DirectoryViewType:
		return "directory", AnalyticsContextBuilder
	case StatusViewType:
		return "status", AnalyticsContextNeutral
	case RedirectViewType:
		return "redirect", AnalyticsContextNeutral
	}
	switch mode {
	case ViewModeHome:
		return "home", AnalyticsContextNeutral
	case ViewModeUser:
		return "user", AnalyticsContextBuilder
	case ViewModePackage:
		return "package", AnalyticsContextBuilder
	case ViewModeRealm:
		return "realm", AnalyticsContextNeutral
	case ViewModeExplorer:
		return "explorer", AnalyticsContextBuilder
	}
	return "other", AnalyticsContextNeutral
}
