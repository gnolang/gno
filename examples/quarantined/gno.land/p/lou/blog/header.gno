package blog

import (
	"time"

	"gno.land/p/lou/query"
	"gno.land/p/moul/md"
)

// HeaderLink applies to a function that transform the rawPath to make the suggested query changes (when header is interacted with)
// will typically returns a md.Link as string
type HeaderLink func(rawPath string) string

type HeaderPreset struct {
	Label   string // title of header (can be your query value, etc...)
	Default string // default value, if none just ""
	Render  HeaderLink
}

func RenderModeToggle(rawPath string) string {
	return md.Link("⊞ grid", GetGridFmtPath(rawPath)) + " | " +
		md.Link("≔ list", GetListFmtPath(rawPath))
}

func RenderTimeFormat(rawPath string) string {
	time := query.GetQuery("time", rawPath)
	if time == "" {
		time = "relative"
	}
	return md.Link("⧖ "+time, GetTimeFmtPath(rawPath, toggleTimeFormat(time)))
}

func RenderSortAlpha(rawPath string) string {
	sort := query.GetQuery("sort", rawPath)
	order := query.GetQuery("order", rawPath)
	if order == "" {
		order = "desc"
	}
	alphaOrder := "(A-Z)"
	if order == "desc" && sort == "alpha" {
		alphaOrder = "(Z-A)"
	}
	return md.Link(orderArrow("alpha", sort, order)+" alphabetical "+alphaOrder,
		GetSortByAlphabeticalPath(rawPath, order))
}

func RenderSortRecent(rawPath string) string {
	sort := query.GetQuery("sort", rawPath)
	order := query.GetQuery("order", rawPath)
	if order == "" {
		order = "desc"
	}
	publishedOrder := " recent"
	if order == "desc" && sort == "recent" {
		publishedOrder = " oldest"
	}
	return md.Link(orderArrow("recent", sort, order)+publishedOrder,
		GetSortByPublishedPath(rawPath, order))
}

func RenderSortUpdate(rawPath string) string {
	sort := query.GetQuery("sort", rawPath)
	order := query.GetQuery("order", rawPath)
	if order == "" {
		order = "desc"
	}
	return md.Link(orderArrow("update", sort, order)+" last updated",
		GetSortByUpdatedPath(rawPath, order))
}

func RenderSortCommon(rawPath string) string {
	sort := query.GetQuery("sort", rawPath)
	order := query.GetQuery("order", rawPath)
	if order == "" {
		order = "desc"
	}
	commonOrder := " most common"
	if order == "desc" && sort == "common" {
		commonOrder = " least common"
	}
	return md.Link(orderArrow("common", sort, order)+commonOrder,
		GetSortByCommonPath(rawPath, order))
}

func RenderHeader(rawPath string, presets []HeaderPreset) string {
	out := ""
	for i, preset := range presets {
		if preset.Render != nil {
			out += preset.Render(rawPath)
		} else {
			out += preset.Label + ": " + preset.Default
		}
		if i < len(presets)-1 {
			out += " | "
		}
	}
	return out
}

// TimeRangePreset struct is used for any customisation of the time range feature
type TimeRangePreset struct {
	Label string // title of time range
	Start string // start time in form of string (check compatibility through `tryParseFlexibleDate`)
	End   string // end time ^^
}

var DefaultTimeRanges = []TimeRangePreset{
	{
		"past year",
		time.Now().AddDate(-1, 0, 0).Format("2006-01-02"),
		time.Now().Format("2006-01-02"),
	},
	{
		"this year",
		time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.Now().Location()).Format("2006-01-02"),
		time.Now().Format("2006-01-02"),
	},
	{
		"last 30 days",
		time.Now().AddDate(0, 0, -30).Format("2006-01-02"),
		time.Now().Format("2006-01-02"),
	},
}

func RenderTimeRangeLinks(rawPath string, presets *[]TimeRangePreset) string {
	out := ""
	timePresets := presets
	if presets == nil {
		timePresets = &DefaultTimeRanges
	}
	for i, r := range *timePresets {
		link, _ := query.UpdateQueryFirstValues(rawPath, map[string]string{
			"start": r.Start,
			"end":   r.End,
		})
		out += md.Link(r.Label, link)
		if i < len(*timePresets)-1 {
			out += ", "
		}
	}
	return out
}
