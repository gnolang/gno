package blog

import (
	"net/url"
	"time"
)

type RenderOptions struct {
	Sort           string
	TimeFormat     string
	StartTime      time.Time // start time for filtering
	EndTime        time.Time // end time for filtering
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
	startTime, endTime := parseTimeRange(rawPath)

	return RenderOptions{
		Sort:           sort,
		TimeFormat:     timeFmt,
		StartTime:      startTime,
		EndTime:        endTime,
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

func GetTimeStartPath(rawPath, startTime string) string {
	return updateQuery(rawPath, "start", startTime)
}

func GetTimeEndPath(rawPath, endTime string) string {
	return updateQuery(rawPath, "end", endTime)
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

func GetSortByCommonPath(rawPath, order string) string {
	return updateQueryMulti(rawPath, map[string]string{
		"sort":  "common",
		"order": toggle(order),
	})
}

func toggle(current string) string {
	if current == "desc" {
		return "asc"
	}
	return "desc"
}

func parseTimeRange(rawPath string) (time.Time, time.Time) {
	startStr := parseQueryParams("start", rawPath)
	endStr := parseQueryParams("end", rawPath)

	var start, end time.Time
	if startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			start = t
		}
	}
	if endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			end = t
		}
	}
	return start, end
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
