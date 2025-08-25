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
}

func StaticHeaderGeneralLinks() []HeaderLink {
	return []HeaderLink{
		{Label: "About", URL: "https://gno.land/about"},
		{Label: "Docs", URL: "https://docs.gno.land/"},
		{Label: "GitHub", URL: "https://github.com/gnolang"},
	}
}

func StaticHeaderDevLinks(u weburl.GnoURL, mode ViewMode) []HeaderLink {
	contentURL, sourceURL, helpURL := u, u, u
	contentURL.WebQuery = url.Values{}
	sourceURL.WebQuery = url.Values{"source": {""}}
	helpURL.WebQuery = url.Values{"help": {""}}

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

	switch mode {
	case ViewModeExplorer:
		return []HeaderLink{}
	case ViewModeUser:
		return []HeaderLink{contentLink}
	case ViewModePackage:
		return []HeaderLink{contentLink, sourceLink}
	default:
		return []HeaderLink{contentLink, sourceLink, actionsLink}
	}
}

func EnrichHeaderData(data HeaderData, mode ViewMode) HeaderData {
	data.RealmPath = data.RealmURL.EncodeURL()
	data.Links.Dev = StaticHeaderDevLinks(data.RealmURL, mode)
	data.Links.General = nil

	if mode.ShouldShowGeneralLinks() {
		data.Links.General = StaticHeaderGeneralLinks()
	}

	return data
}

func isActive(webQuery url.Values, label string) bool {
	switch label {
	case "Content":
		return !webQuery.Has("source") && !webQuery.Has("help")
	case "Source":
		return webQuery.Has("source")
	case "Actions":
		return webQuery.Has("help")
	default:
		return false
	}
}
