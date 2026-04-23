package components

// Outbound* labels are emitted as data-outbound on tagged links so the
// SimpleAnalytics auto-events script fires outbound_<label> events. The set
// must stay in sync with the enum in frontend/js/analytics.ts and the values
// documented in SIMPLEANALYTICS.md.
const (
	OutboundDocs    = "docs"
	OutboundFaucet  = "faucet"
	OutboundStatus  = "status"
	OutboundGitHub  = "github"
	OutboundTwitter = "twitter"
	OutboundDiscord = "discord"
	OutboundYouTube = "youtube"
)

// AnalyticsData holds the SimpleAnalytics metadata rendered into the page.
type AnalyticsData struct {
	Enabled    bool
	PageType   string
	ChainId    string
	AssetsPath string
	BuildTime  string
	// Hostname, when non-empty, is rendered as data-hostname on the
	// SimpleAnalytics script tag to override the hostname SA reports.
	// Set this when the site listens on a host SA would otherwise report
	// incorrectly (for example a non-default port in local development).
	Hostname string
}

// ClassifyPageType returns the page-type label for a given mode and view.
// View type takes precedence over mode: a Source view inside a Realm mode is
// classified as "source", not "realm", so the analytics label matches the
// rendered surface rather than the containing layout mode.
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
		return "pure"
	case ViewModeRealm:
		return "realm"
	case ViewModeExplorer:
		return "explorer"
	}
	return "other"
}
