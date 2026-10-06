package components

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

type HeaderLink struct {
	Label    string
	URL      string
	Icon     string
	IsActive bool
	// Tooltip is shown on hover to explain what the link opens. Empty renders
	// no title attribute.
	Tooltip string
	// Outbound, when set to one of the Outbound* constants, is rendered as
	// data-outbound on the link so SimpleAnalytics fires a named
	// outbound_<label> event instead of an anonymous outbound click.
	Outbound string
}

type HeaderLinks struct {
	General []HeaderLink
	Dev     []HeaderLink
}

type HeaderData struct {
	RealmPath string
	// SearchAction is the omnibar form's target. It exists so the bar works
	// with JavaScript disabled: the form submits here and the reader gets
	// the server-rendered results page. With JavaScript the controller
	// intercepts the submit, so this stays the fallback rather than the
	// normal path.
	SearchAction string
	RealmURL     weburl.GnoURL
	Breadcrumb   BreadcrumbData
	Links        HeaderLinks
	ChainId      string
	Remote       string
	Mode         ViewMode
	Static       bool
	// Origin is the request scheme+host the AI prompts link to.
	Origin string
	AI     *AIMenu
	Notice RealmNotice
	// Listing is set on a directory listing, whose tabs it carries.
	Listing *ListingTabs
}

// RealmNotice is the header row shown on pages of community packages.
type RealmNotice struct {
	// Text is shown at every width, unless Short is set.
	Text BannerData
	// Short, if set, replaces Text below the lg breakpoint.
	Short BannerData
}

// Enabled reports whether the page shows the notice: handlers only set a
// RealmNotice on pages of packages outside the trusted paths.
func (n RealmNotice) Enabled() bool { return n.Text.Enabled() }

// Lines is how many lines the row reserves in the sticky header, 0 when
// disabled. A notice with a short variant is the default one, whose texts fit
// one line at every width; any other text is clamped to two.
func (n RealmNotice) Lines() int {
	switch {
	case !n.Enabled():
		return 0
	case n.Short.Enabled():
		return 1
	default:
		return 2
	}
}

// NewRealmNotice renders text and short as inline markdown, the same way as
// NewBannerData but without images; short may be empty. Unlike the opt-in
// banner, a text that shows no visible character is an error, so the notice
// cannot switch itself off on a typo or be blanked on purpose.
func NewRealmNotice(text, short string) (RealmNotice, error) {
	t, err := newNoticeText(text)
	if err != nil {
		return RealmNotice{}, err
	}
	n := RealmNotice{Text: t}
	if short != "" {
		if n.Short, err = newNoticeText(short); err != nil {
			return RealmNotice{}, fmt.Errorf("short variant: %w", err)
		}
	}
	return n, nil
}

func newNoticeText(markdown string) (BannerData, error) {
	b, visible, err := renderInline(markdown, "", true)
	if err != nil {
		return BannerData{}, err
	}
	if !visible {
		return BannerData{}, errors.New("renders no visible text")
	}
	return b, nil
}

func StaticHeaderGeneralLinks() []HeaderLink {
	return []HeaderLink{
		{Label: "About", URL: "https://gno.land/about"},
		{Label: "Docs", URL: "https://docs.gno.land/", Outbound: OutboundDocs},
		{Label: "GitHub", URL: "https://github.com/gnolang", Outbound: OutboundGitHub},
	}
}

func StaticHeaderDevLinks(u weburl.GnoURL, mode ViewMode, static bool) []HeaderLink {
	contentURL, sourceURL, helpURL, stateURL := u, u, u, u
	contentURL.WebQuery = url.Values{}
	sourceURL.WebQuery = url.Values{"source": {""}}
	helpURL.WebQuery = url.Values{"help": {""}}
	stateURL.WebQuery = url.Values{"state": {""}}

	// Carry the open file onto the Source link. Without it, a reader already
	// looking at a file is sent back to the package overview by the tab that
	// is meant to be showing them source.
	if file := u.WebQuery.Get("file"); file != "" {
		sourceURL.WebQuery.Set("file", file)
	}

	contentLink := HeaderLink{
		Label:    "Content",
		URL:      contentURL.EncodeWebURL(),
		Icon:     "ico-content",
		IsActive: isActive(u.WebQuery, "Content"),
		Tooltip:  "The realm's rendered page, or a package's file listing.",
	}

	sourceLink := HeaderLink{
		Label:    "Source",
		URL:      sourceURL.EncodeWebURL(),
		Icon:     "ico-code",
		IsActive: isActive(u.WebQuery, "Source"),
		Tooltip:  "Browse the package source files.",
	}

	actionsLink := HeaderLink{
		Label:    "Actions",
		URL:      helpURL.EncodeWebURL(),
		Icon:     "ico-helper",
		IsActive: isActive(u.WebQuery, "Actions"),
		Tooltip:  "Call the realm's exported functions.",
	}

	stateLink := HeaderLink{
		Label:    "State",
		URL:      stateURL.EncodeWebURL(),
		Icon:     "ico-state",
		IsActive: isActive(u.WebQuery, "State"),
		Tooltip:  "Inspect the realm's stored on-chain state.",
	}

	switch {
	case static:
		return []HeaderLink{contentLink}
	case mode == ViewModeExplorer:
		return []HeaderLink{}
	case mode == ViewModeUser:
		return []HeaderLink{contentLink}
	case mode == ViewModePackage:
		return []HeaderLink{contentLink, sourceLink}
	default:
		return []HeaderLink{contentLink, stateLink, sourceLink, actionsLink}
	}
}

// ListingTabs are the renderings a directory listing offers, set by the
// listing page itself: nothing else in explorer mode is a listing.
type ListingTabs struct {
	// Map is false for a listing too small for a map to show anything a list
	// does not.
	Map bool
}

// links are the Directory and Map tabs, or none when a map is not offered: a
// single tab would choose between nothing.
func (t *ListingTabs) links(u weburl.GnoURL) []HeaderLink {
	if t == nil || !t.Map {
		return nil
	}
	listURL, mapURL := u, u
	listURL.WebQuery = url.Values{}
	mapURL.WebQuery = url.Values{"map": {""}}
	onMap := u.WebQuery.Has("map")
	return []HeaderLink{
		{
			Label:    "Directory",
			URL:      listURL.EncodeWebURL(),
			Icon:     "ico-folder",
			IsActive: !onMap,
			Tooltip:  "Every package under this path, one per line.",
		},
		{
			Label:    "Map",
			URL:      mapURL.EncodeWebURL(),
			Icon:     "ico-grid",
			IsActive: onMap,
			Tooltip:  "The same packages drawn as a map, grouped by path.",
		},
	}
}

func EnrichHeaderData(data HeaderData, mode ViewMode) HeaderData {
	data.RealmPath = data.RealmURL.EncodeURL()
	// A page that names no package still searches — chain-wide.
	searchBase := data.RealmURL.Path
	if searchBase == "" {
		searchBase = "/"
	}
	data.SearchAction = searchBase + "$search"
	data.Links.Dev = StaticHeaderDevLinks(data.RealmURL, mode, data.Static)
	if mode == ViewModeExplorer {
		data.Links.Dev = data.Listing.links(data.RealmURL)
	}
	if !data.Static && (mode == ViewModeRealm || mode == ViewModePackage) {
		data.AI = NewAIMenu(data.Origin, data.RealmURL)
	}
	data.Links.General = nil

	if mode.ShouldShowGeneralLinks() {
		data.Links.General = StaticHeaderGeneralLinks()
	}

	return data
}

func isActive(webQuery url.Values, label string) bool {
	switch label {
	case "Content":
		return !webQuery.Has("source") && !webQuery.Has("help") && !webQuery.Has("state") && !webQuery.Has("deps")
	case "State":
		return webQuery.Has("state")
	case "Source":
		// The dependencies page extends the overview, which is under Source.
		return webQuery.Has("source") || webQuery.Has("deps")
	case "Actions":
		return webQuery.Has("help")
	default:
		return false
	}
}
