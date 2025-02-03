package components

import (
	"net/url"
	"strings"
)

type HeaderLink struct {
	Label    string
	URL      string
	Icon     string
	IsActive bool
}

type HeaderData struct {
	RealmPath  string
	Breadcrumb BreadcrumbData
	WebQuery   url.Values
	Links      []HeaderLink
}

func HeaderURL(realmPath string, webquery string) string {
	if webquery == "" {
		return realmPath
	}
	if pos := strings.IndexByte(realmPath, '?'); pos != -1 {
		return realmPath[:pos] + webquery + realmPath[pos:]
	}
	return realmPath + webquery
}

func StaticHeaderLinks(realmPath string, webQuery url.Values) []HeaderLink {
	return []HeaderLink{
		{
			Label:    "Content",
			URL:      HeaderURL(realmPath, ""),
			Icon:     "ico-info",
			IsActive: isActive(webQuery, "Content"),
		},
		{
			Label:    "Source",
			URL:      HeaderURL(realmPath, "$source"),
			Icon:     "ico-code",
			IsActive: isActive(webQuery, "Source"),
		},
		{
			Label:    "Docs",
			URL:      HeaderURL(realmPath, "$help"),
			Icon:     "ico-docs",
			IsActive: isActive(webQuery, "Docs"),
		},
	}
}

func EnrichHeaderData(data HeaderData) HeaderData {
	data.Links = StaticHeaderLinks(data.RealmPath, data.WebQuery)
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
