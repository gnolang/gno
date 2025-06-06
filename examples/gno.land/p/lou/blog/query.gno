package blog

import "net/url"

type RenderOptions struct {
	Sort       string
	TimeFormat string
	// add start and current time to options
	Ascending      bool
	NewestFirst    bool
	IsGrid         bool
	IsAlphabetical bool
	IsLastUpdated  bool
	IsCommon       bool
	HasOrder       bool
}

func parseRenderOptions(rawPath string) RenderOptions {
	sort := parseQueryParams("sort", rawPath)
	order := parseQueryParams("order", rawPath)
	mode := parseQueryParams("mode", rawPath)
	timeFmt := parseQueryParams("time", rawPath)

	return RenderOptions{
		Sort:           sort,
		TimeFormat:     timeFmt,
		NewestFirst:    sort == "recent" || sort == "",
		Ascending:      order == "asc" || order == "",
		IsGrid:         mode == "grid" || mode == "",
		IsAlphabetical: sort == "alpha",
		IsLastUpdated:  sort == "update",
		IsCommon:       sort == "common",
		HasOrder:       order == "asc" || order == "desc",
	}
}

func GetGridFmtPath(rawPath string) string {
	return updateQuery(rawPath, "mode", "grid")
}

func GetListFmtPath(rawPath string) string {
	return updateQuery(rawPath, "mode", "list")
}

func GetSortByPublishedPath(rawPath, order string) string { // change to use toggle?
	if order == "asc" {
		return updateQuery(rawPath, "sort", "recent")
	}
	return updateQuery(rawPath, "sort", "oldest")
}

func GetTimeFmtPath(rawPath, timeFormat string) string {
	return updateQuery(rawPath, "time", timeFormat)
}

func GetSortByUpdatedPath(rawPath, order string) string {
	return updateQueryMulti(rawPath, map[string]string{
		"sort":  "update",
		"order": toggle(order),
	})
}

func GetSortByAlphabeticalPath(rawPath, order string) string {
	return updateQueryMulti(rawPath, map[string]string{
		"sort":  "alpha",
		"order": toggle(order),
	})
}

func GetSortByCommonPath(rawPath, order string) string { // change to use toggle?
	return updateQueryMulti(rawPath, map[string]string{
		"sort":  "common",
		"order": order,
	})
}

func toggle(current string) string {
	if current == "desc" {
		return "asc"
	}
	return "desc"
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
