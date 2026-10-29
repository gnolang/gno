package components

import (
	"errors"
	"fmt"
	"net/url"
	"slices"

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
	// Path is the plain path a menu item shows under its label, where URL
	// may carry a tab and an anchor ($source#subpackages).
	Path string
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
	// MapTab is set by a directory listing large enough for a map: it then
	// offers its Directory and Map renderings as tabs.
	MapTab bool
	// CopyOf is the package a playground copy was opened from, whose tabs
	// the copy shows, none of them current.
	CopyOf *weburl.GnoURL
	// Error is set on an error page. Its tabs leave out Run, whose page
	// needs a package that exists.
	Error bool
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
	contentURL, sourceURL, helpURL, runURL, stateURL := u, u, u, u, u
	contentURL.WebQuery = url.Values{}
	sourceURL.WebQuery = url.Values{"source": {""}}
	helpURL.WebQuery = url.Values{"help": {""}}
	runURL.WebQuery = url.Values{"run": {""}}
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

	runLink := HeaderLink{
		Label:    "Run",
		URL:      runURL.EncodeWebURL(),
		Icon:     "ico-tx-link",
		IsActive: isActive(u.WebQuery, "Run"),
		Tooltip:  "Write a script that calls the realm and get its gnokey command.",
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
	case mode == ViewModePlayground:
		// The breadcrumb's playground pill says where the reader is.
		return []HeaderLink{}
	default:
		return []HeaderLink{contentLink, stateLink, sourceLink, actionsLink, runLink}
	}
}

// listingTabs are a listing's Directory and Map tabs.
func listingTabs(u weburl.GnoURL) []HeaderLink {
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
	// The root has no path: an empty value shows the placeholder
	if data.RealmURL.Path != "/" {
		data.RealmPath = data.RealmURL.EncodeURL()
	}
	// A page that names no package still searches — chain-wide.
	searchBase := data.RealmURL.Path
	if searchBase == "" {
		searchBase = "/"
	}
	data.SearchAction = searchBase + "$search"
	switch {
	case data.MapTab:
		data.Links.Dev = listingTabs(data.RealmURL)
	case mode == ViewModeExplorer:
		// A listing without a map has one rendering: a single tab would
		// choose between nothing.
		data.Links.Dev = nil
	case mode == ViewModePlayground && data.CopyOf != nil:
		copyMode := ViewModeRealm
		if data.CopyOf.IsPure() {
			copyMode = ViewModePackage
		}
		data.Links.Dev = StaticHeaderDevLinks(*data.CopyOf, copyMode, false)
		for i := range data.Links.Dev {
			data.Links.Dev[i].IsActive = false
		}
	default:
		data.Links.Dev = StaticHeaderDevLinks(data.RealmURL, mode, data.Static)
		if data.Error {
			data.Links.Dev = slices.DeleteFunc(data.Links.Dev, func(l HeaderLink) bool { return l.Label == "Run" })
		}
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
		return !webQuery.Has("source") && !webQuery.Has("help") && !webQuery.Has("run") && !webQuery.Has("state") && !webQuery.Has("deps")
	case "State":
		return webQuery.Has("state")
	case "Source":
		// The dependencies page extends the overview, which is under Source.
		return webQuery.Has("source") || webQuery.Has("deps")
	case "Actions":
		return webQuery.Has("help")
	case "Run":
		return webQuery.Has("run")
	default:
		return false
	}
}
