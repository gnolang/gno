package blog

import (
	"net/url"
	"strings"

	"gno.land/p/moul/md"
)

func renderTagLinks(prefix string, tags []string) string {
	return renderLinks(prefix, "tag", tags, "#")
}

func renderAuthorLinks(prefix string, authors []string) string {
	return renderLinks(prefix, "author", authors, "")
}

func renderLinks(prefix, kind string, values []string, labelPrefix string) string {
	links := make([]string, len(values))
	for i, val := range values {
		label := labelPrefix + val
		href := prefix + ":" + kind + "/" + val
		links[i] = md.Link(label, href)
	}
	return strings.Join(links, ", ") + "\n"
}

func parseQueryParams(query, rawPath string) string {
	u, err := url.Parse(rawPath)
	if err != nil {
		return "grid"
	}
	return u.Query().Get(query)
}

func hasField(fields []string, match string) bool {
	for _, field := range fields {
		if field == match {
			return true
		}
	}
	return false
}

func updateQuery(rawPath, key, value string) string {
	u, err := url.Parse(rawPath)
	if err != nil {
		return rawPath
	}
	query := u.Query()
	query.Set(key, value)
	u.RawQuery = query.Encode()
	return u.String()
}
