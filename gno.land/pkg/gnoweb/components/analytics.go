package components

// AnalyticsData holds the SimpleAnalytics metadata rendered into the page.
type AnalyticsData struct {
	Enabled    bool
	PageType   string
	ChainId    string
	AssetsPath string
	BuildTime  string
}

// ClassifyPageType returns the page-type label for a given mode and view.
// View type takes precedence when both match: a Source view inside a Realm
// mode is "source", not "realm".
func ClassifyPageType(mode ViewMode, view ViewType) string {
	switch view {
	case SourceViewType:
		return "source"
	case HelpViewType:
		return "help"
	case DirectoryViewType:
		return "directory"
	case StatusViewType:
		return "status"
	case RedirectViewType:
		return "redirect"
	}
	switch mode {
	case ViewModeHome:
		return "home"
	case ViewModeUser:
		return "user"
	case ViewModePackage:
		return "package"
	case ViewModeRealm:
		return "realm"
	case ViewModeExplorer:
		return "explorer"
	}
	return "other"
}
