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
	HasDevMenu bool
}

func StaticHeaderGeneralLinks() []HeaderLink {
	links := []HeaderLink{
		{
			Label: "About",
			URL:   "https://gno.land/about",
		},
		{
			Label: "Docs",
			URL:   "https://docs.gno.land/",
		},
		{
			Label: "GitHub",
			URL:   "https://github.com/gnolang",
		},
	}
	return links
}

func StaticHeaderDevLinks(u weburl.GnoURL, handle string) []HeaderLink {
	contentURL, sourceURL, helpURL := u, u, u
	contentURL.WebQuery = url.Values{}
	sourceURL.WebQuery = url.Values{"source": {""}}
	helpURL.WebQuery = url.Values{"help": {""}}

	links := []HeaderLink{
		{
			Label:    "Content",
			URL:      contentURL.EncodeWebURL(),
			Icon:     "ico-content",
			IsActive: isActive(u.WebQuery, "Content"),
		},
		{
			Label:    "Source",
			URL:      sourceURL.EncodeWebURL(),
			Icon:     "ico-code",
			IsActive: isActive(u.WebQuery, "Source"),
		},
	}

	switch handle {
	case "p":
		// Will have docs soon

	default:
		links = append(links, HeaderLink{
			Label:    "Actions",
			URL:      helpURL.EncodeWebURL(),
			Icon:     "ico-helper",
			IsActive: isActive(u.WebQuery, "Actions"),
		})
	}

	return links
}

func EnrichHeaderData(data HeaderData, hasDevMenu bool) HeaderData {
	data.RealmPath = data.RealmURL.EncodeURL()

	var handle string
	if len(data.Breadcrumb.Parts) > 0 {
		handle = data.Breadcrumb.Parts[0].Name
	} else {
		handle = ""
	}

	data.Links.Dev = StaticHeaderDevLinks(data.RealmURL, handle)

	if hasDevMenu {
		data.Links.General = nil
	} else {
		data.Links.General = StaticHeaderGeneralLinks()
	}

	return data
}

func isActive(webQuery url.Values, label string) bool {
	switch label {
	case "Content":
		return !(webQuery.Has("source") || webQuery.Has("help"))
	case "Source":
		return webQuery.Has("source")
	case "Actions":
		return webQuery.Has("help")
	default:
		return false
	}
}
