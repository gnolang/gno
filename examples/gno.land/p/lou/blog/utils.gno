package blog

import (
	"net/url"
	"strings"

	"gno.land/p/moul/md"
)

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
		return ""
	}
	return u.Query().Get(query)
}

func parseQueryMap(rawPath string) map[string]string {
	u, err := url.Parse(rawPath)
	if err != nil {
		return map[string]string{}
	}
	params := make(map[string]string)
	for key, vals := range u.Query() {
		if len(vals) > 0 {
			params[key] = vals[0]
		}
	}
	return params
}

func buildQueryPath(rawPath string, params map[string]string) string {
	u, err := url.Parse(rawPath)
	if err != nil {
		return rawPath
	}
	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func updateQueryMulti(rawPath string, updates map[string]string) string {
	params := parseQueryMap(rawPath)
	for k, v := range updates {
		params[k] = v
	}
	return buildQueryPath(rawPath, params)
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

func hasField(fields []string, match string) bool {
	for _, field := range fields {
		if field == match {
			return true
		}
	}
	return false
}

func orderArrow(targetSort, currentSort, order string) string {
	if targetSort != currentSort {
		return "↕"
	}
	if order == "asc" {
		return "▲"
	}
	return "▼"
}

func titleToSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
			return r
		}
		return -1
	}, slug)
	return slug
}

func renderBreadcrumb(prefix, path string) string {
	segments := strings.Split(path, "/")
	base := prefix
	var parts []string

	if prefix != "" {
		parts = append(parts, md.Link(prefix, prefix))
	}

	for i, seg := range segments {
		if seg == "" {
			continue
		}
		base += ":" + seg
		label := seg
		if i == len(segments)-1 {
			label = seg
			parts = append(parts, label)
			break
		}
		parts = append(parts, md.Link(label, base))
	}
	return strings.Join(parts, "/")
}
