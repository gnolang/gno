package timeutil

import (
	"time"

	"gno.land/p/demo/ufmt"
)

func CalculateTimeSince(t time.Time) string {
	duration := time.Since(t)
	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		if duration < 2*time.Minute {
			return "a minute ago"
		}
		return ufmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	case duration < 24*time.Hour:
		if duration < 2*time.Hour {
			return "an hour ago"
		}
		return ufmt.Sprintf("%d hours ago", int(duration.Hours()))
	case duration < 7*24*time.Hour:
		if duration < 2*24*time.Hour {
			return "a day ago"
		}
		return ufmt.Sprintf("%d days ago", int(duration.Hours()/24))
	case duration < 30*24*time.Hour:
		if duration < 2*7*24*time.Hour {
			return "a week ago"
		}
		return ufmt.Sprintf("%d weeks ago", int(duration.Hours()/(7*24)))
	default:
		if duration < 2*30*24*time.Hour {
			return "a month ago"
		}
		return ufmt.Sprintf("%d months ago", int(duration.Hours()/(30*24)))
	}
}

func TryParseFlexibleDate(s string) *time.Time {
	formats := []string{
		time.RFC3339, // rfc3339 format
		"2006-01-02", // date only
		"2006-01",    // year and month
		"2006",       // year only
	}
	return ParseFlexibleDate(s, formats)
}

func ParseFlexibleDate(s string, formats []string) *time.Time {
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return &t
		}
	}
	return nil
}
