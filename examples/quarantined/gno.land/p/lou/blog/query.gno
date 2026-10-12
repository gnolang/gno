package blog

import (
	"time"

	"gno.land/p/lou/query"
)

type RenderOptions struct {
	Sort           string
	TimeFormat     string
	StartTime      *time.Time
	EndTime        *time.Time
	Ascending      bool
	IsGrid         bool
	IsAlphabetical bool
	IsLastUpdated  bool
	IsCommon       bool
	HasOrder       bool
}

func ParseRenderOptions(rawPath string) RenderOptions {
	sort := query.GetQuery("sort", rawPath)
	order := query.GetQuery("order", rawPath)
	mode := query.GetQuery("mode", rawPath)
	timeFmt := query.GetQuery("time", rawPath)
	startTime, endTime := parseTimeRange(rawPath)

	return RenderOptions{
		Sort:           sort,
		TimeFormat:     timeFmt,
		StartTime:      startTime,
		EndTime:        endTime,
		Ascending:      order == "asc" || order == "",
		IsGrid:         mode == "grid" || mode == "",
		IsAlphabetical: sort == "alpha",
		IsLastUpdated:  sort == "update",
		IsCommon:       sort == "common",
		HasOrder:       order == "asc" || order == "desc",
	}
}

func GetGridFmtPath(rawPath string) string {
	link, _ := query.UpdateQueryValue(rawPath, "mode", "grid")
	return link
}

func GetListFmtPath(rawPath string) string {
	link, _ := query.UpdateQueryValue(rawPath, "mode", "list")
	return link
}

func GetSortByPublishedPath(rawPath, order string) string {
	link, _ := query.UpdateQueryFirstValues(rawPath, map[string]string{
		"sort":  "recent",
		"order": toggle(order),
	})
	return link
}

func GetTimeFmtPath(rawPath, timeFormat string) string {
	link, _ := query.UpdateQueryValue(rawPath, "time", timeFormat)
	return link
}

func GetTimeStartPath(rawPath, startTime string) string {
	link, _ := query.UpdateQueryValue(rawPath, "start", startTime)
	return link
}

func GetTimeEndPath(rawPath, endTime string) string {
	link, _ := query.UpdateQueryValue(rawPath, "end", endTime)
	return link
}

func GetSortByUpdatedPath(rawPath, order string) string {
	link, _ := query.UpdateQueryFirstValues(rawPath, map[string]string{
		"sort":  "update",
		"order": toggle(order),
	})
	return link
}

func GetSortByAlphabeticalPath(rawPath, order string) string {
	link, _ := query.UpdateQueryFirstValues(rawPath, map[string]string{
		"sort":  "alpha",
		"order": toggle(order),
	})
	return link
}

func GetSortByCommonPath(rawPath, order string) string {
	link, _ := query.UpdateQueryFirstValues(rawPath, map[string]string{
		"sort":  "common",
		"order": toggle(order),
	})
	return link
}

func toggle(current string) string {
	if current == "desc" {
		return "asc"
	}
	return "desc"
}

func toggleTimeFormat(current string) string {
	if current == "relative" {
		return "short"
	}
	if current == "short" {
		return "full"
	}
	if current == "full" {
		return "relative"
	}
	return "short"
}

func parseTimeRange(rawPath string) (*time.Time, *time.Time) {
	startStr := query.GetQuery("start", rawPath)
	endStr := query.GetQuery("end", rawPath)
	var start, end *time.Time

	if startStr != "" {
		if t := tryParseFlexibleDate(startStr); t != nil {
			start = t
		}
	}
	if endStr != "" {
		if t := tryParseFlexibleDate(endStr); t != nil {
			end = t
		}
	}
	return start, end
}
