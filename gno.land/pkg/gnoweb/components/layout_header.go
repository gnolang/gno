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

type HeaderData struct {
	RealmPath  string
	RealmURL   weburl.GnoURL
	Breadcrumb BreadcrumbData
	Links      []HeaderLink
	ChainId    string
	Remote     string
}

func StaticHeaderLinks(u weburl.GnoURL) []HeaderLink {
	contentURL, sourceURL, helpURL := u, u, u
	contentURL.WebQuery = url.Values{}
	sourceURL.WebQuery = url.Values{"source": {""}}
	helpURL.WebQuery = url.Values{"help": {""}}

	return []HeaderLink{
		{
			Label:    "Content",
			URL:      contentURL.EncodeWebURL(),
			Icon:     "ico-info",
			IsActive: isActive(u.WebQuery, "Content"),
		},
		{
			Label:    "Source",
			URL:      sourceURL.EncodeWebURL(),
			Icon:     "ico-code",
			IsActive: isActive(u.WebQuery, "Source"),
		},
		{
			Label:    "Docs",
			URL:      helpURL.EncodeWebURL(),
			Icon:     "ico-docs",
			IsActive: isActive(u.WebQuery, "Docs"),
		},
	}
}

func EnrichHeaderData(data HeaderData) HeaderData {
	data.RealmPath = data.RealmURL.EncodeURL()
	data.Links = StaticHeaderLinks(data.RealmURL)
	return data
}

func isActive(webQuery url.Values, label string) bool {
	switch label {
	case "Content":
		return !(webQuery.Has("source") || webQuery.Has("help"))
	case "Source":
		return webQuery.Has("source")
	case "Docs":
		return webQuery.Has("help")
	default:
		return false
	}
}
