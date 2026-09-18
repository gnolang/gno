package components

import (
	"net/url"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

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
	Enabled  bool
	PageType string
	// Path is the analytics pageview path (see analyticsPath) rendered as
	// data-sa-path; the client reports it to SimpleAnalytics in place of the
	// raw pathname.
	Path       string
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

// analyticsRedacted is the placeholder analyticsPath substitutes for masked
// web-query values.
const analyticsRedacted = "redacted"

// analyticsWebQueryValueKeys lists the web-query keys whose values analyticsPath
// reports verbatim; every other key's value is masked.
var analyticsWebQueryValueKeys = map[string]bool{
	"func": true, // function name
	"file": true, // source file name
}

// analyticsPath returns the path reported to SimpleAnalytics via its
// path-overwriter hook, in place of the raw pathname. It keeps the route path,
// render args, and the func and file web-query values verbatim; valueless flags
// (e.g. help, source) survive because empty values are not masked. Every other
// web-query value is masked to analyticsRedacted, and the standard query is
// dropped. The route and render args are retained by design, so data a realm
// encodes there still reaches SimpleAnalytics.
func analyticsPath(u weburl.GnoURL) string {
	if len(u.WebQuery) > 0 {
		masked := make(url.Values, len(u.WebQuery))
		for key, values := range u.WebQuery {
			if analyticsWebQueryValueKeys[key] {
				masked[key] = values
				continue
			}
			redacted := make([]string, len(values))
			for i, v := range values {
				if v != "" { // preserve valueless mode flags
					redacted[i] = analyticsRedacted
				}
			}
			masked[key] = redacted
		}
		u.WebQuery = masked
	}
	u.Query = nil

	encoded := u.EncodeWebURL()
	if encoded == "" {
		return "/"
	}
	return encoded
}
