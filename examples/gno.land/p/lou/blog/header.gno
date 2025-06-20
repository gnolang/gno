package blog

import (
	"time"

	"gno.land/p/mason/md"
)

type TimeRangePreset struct {
	Label string
	Start string
	End   string
}

var TimeRangePresets = []TimeRangePreset{
	{"past year", "2024-01-01", "2024-12-31"},
	{"this year", "2025-01-01", "2025-12-31"},
	{"last 30 days", time.Now().AddDate(0, 0, -30).Format("2006-01-02"), time.Now().Format("2006-01-02")},
}

func renderTimeRangeLinks(rawPath string) string {
	var out string
	for i, r := range TimeRangePresets {
		out += md.Link(r.Label, updateQueryMulti(rawPath, map[string]string{
			"start": r.Start,
			"end":   r.End,
		}))
		if i < len(TimeRangePresets)-1 {
			out += ", "
		}
	}
	return out
}
