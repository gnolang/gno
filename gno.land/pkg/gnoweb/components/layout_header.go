package components

import (
	"net/url"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

type HeaderLink struct {
	Label    string
	URL      string
	Icon     string
	IsActive bool
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
}

func StaticHeaderGeneralLinks() []HeaderLink {
	return []HeaderLink{
		{Label: "About", URL: "https://gno.land/about"},
		{Label: "Docs", URL: "https://docs.gno.land/"},
		{Label: "GitHub", URL: "https://github.com/gnolang"},
	}
}

func StaticHeaderDevLinks(u weburl.GnoURL, mode ViewMode, static bool) []HeaderLink {
	contentURL, sourceURL, helpURL, evalURL, forkURL, runURL := u, u, u, u, u, u
	contentURL.WebQuery = url.Values{}
	sourceURL.WebQuery = url.Values{"source": {""}}
	helpURL.WebQuery = url.Values{"help": {""}}
	evalURL.WebQuery = url.Values{"eval": {""}}
	forkURL.WebQuery = url.Values{"fork": {""}}
	runURL.WebQuery = url.Values{"run": {""}}

	contentLink := HeaderLink{
		Label:    "Content",
		URL:      contentURL.EncodeWebURL(),
		Icon:     "ico-content",
		IsActive: isActive(u.WebQuery, "Content"),
	}

	sourceLink := HeaderLink{
		Label:    "Source",
		URL:      sourceURL.EncodeWebURL(),
		Icon:     "ico-code",
		IsActive: isActive(u.WebQuery, "Source"),
	}

	actionsLink := HeaderLink{
		Label:    "Actions",
		URL:      helpURL.EncodeWebURL(),
		Icon:     "ico-helper",
		IsActive: isActive(u.WebQuery, "Actions"),
	}

	evalLink := HeaderLink{
		Label:    "Eval",
		URL:      evalURL.EncodeWebURL(),
		Icon:     "ico-tx-link",
		IsActive: isActive(u.WebQuery, "Eval"),
	}

	forkLink := HeaderLink{
		Label:    "Fork",
		URL:      forkURL.EncodeWebURL(),
		Icon:     "ico-link",
		IsActive: isActive(u.WebQuery, "Fork"),
	}

	runLink := HeaderLink{
		Label:    "Run",
		URL:      runURL.EncodeWebURL(),
		Icon:     "ico-tx-link",
		IsActive: isActive(u.WebQuery, "Run"),
	}

	switch {
	case static:
		return []HeaderLink{contentLink}
	case mode == ViewModeExplorer:
		return []HeaderLink{}
	case mode == ViewModeUser:
		return []HeaderLink{contentLink}
	case mode == ViewModePackage:
		return []HeaderLink{contentLink, sourceLink, forkLink}
	case mode == ViewModePlayground:
		return []HeaderLink{}
	default:
		return []HeaderLink{contentLink, sourceLink, actionsLink, evalLink, forkLink, runLink}
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
		return !webQuery.Has("source") && !webQuery.Has("help") && !webQuery.Has("eval") && !webQuery.Has("fork") && !webQuery.Has("run")
	case "Source":
		return webQuery.Has("source")
	case "Actions":
		return webQuery.Has("help")
	case "Eval":
		return webQuery.Has("eval")
	case "Fork":
		return webQuery.Has("fork")
	case "Run":
		return webQuery.Has("run")
	default:
		return false
	}
}
