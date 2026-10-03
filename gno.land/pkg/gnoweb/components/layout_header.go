package components

import (
	"errors"
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
	RealmPath  string
	RealmURL   weburl.GnoURL
	Breadcrumb BreadcrumbData
	Links      HeaderLinks
	ChainId    string
	Remote     string
	Mode       ViewMode
	Static     bool
	Notice     RealmNotice
}

// RealmNotice is the header row shown on pages of community packages.
type RealmNotice struct {
	// Text is shown at every width, unless Short is set.
	Text BannerData
	// Short, if set, replaces Text below the md breakpoint.
	Short BannerData
}

// Enabled reports whether the page shows the notice: handlers only set a
// RealmNotice on pages of packages outside the trusted paths.
func (n RealmNotice) Enabled() bool { return n.Text.Enabled() }

// NewRealmNotice renders text and short as inline markdown, the same way as
// NewBannerData; short may be empty. Unlike the opt-in banner, a text that
// renders to nothing is an error, so the notice cannot switch itself off on
// a typo.
func NewRealmNotice(text, short string) (RealmNotice, error) {
	t, err := NewBannerData(text, "")
	if err != nil {
		return RealmNotice{}, err
	}
	if !t.Enabled() {
		return RealmNotice{}, errors.New("renders to nothing")
	}
	s, err := NewBannerData(short, "")
	if err != nil {
		return RealmNotice{}, err
	}
	return RealmNotice{Text: t, Short: s}, nil
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

func EnrichHeaderData(data HeaderData, mode ViewMode) HeaderData {
	data.RealmPath = data.RealmURL.EncodeURL()
	data.Links.Dev = StaticHeaderDevLinks(data.RealmURL, mode, data.Static)
	data.Links.General = nil

	if mode.ShouldShowGeneralLinks() {
		data.Links.General = StaticHeaderGeneralLinks()
	}

	return data
}

func isActive(webQuery url.Values, label string) bool {
	switch label {
	case "Content":
		return !webQuery.Has("source") && !webQuery.Has("help") && !webQuery.Has("state")
	case "State":
		return webQuery.Has("state")
	case "Source":
		return webQuery.Has("source")
	case "Actions":
		return webQuery.Has("help")
	default:
		return false
	}
}
