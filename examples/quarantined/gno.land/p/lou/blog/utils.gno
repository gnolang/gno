package blog

import (
	"net/url"
	"strings"
	"time"

	"gno.land/p/moul/md"
	"gno.land/p/nt/ufmt/v0"
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

func hasField(fields []string, match string) bool {
	for _, field := range fields {
		if field == match {
			return true
		}
	}
	return false
}

func extractListingDisplay(key string, value interface{}) (label string, count int) {
	if strings.Contains(key, "::") {
		parts := strings.Split(key, "::")
		if len(parts) > 1 {
			return parts[1], value.(int)
		}
	}
	return key, value.(int)
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

func formatTime(t time.Time, format string) string {
	switch format {
	case "relative":
		return calculateTimeSince(t)
	case "short":
		return t.Format("02 Jan 2006")
	case "full":
		return t.Format("January 02, 2006 at 3:04 PM")
	default:
		return calculateTimeSince(t)
	}
}

func calculateTimeSince(t time.Time) string {
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

func tryParseFlexibleDate(s string) *time.Time {
	formats := []string{
		time.RFC3339, // rfc3339 format
		"2006-01-02", // date only
		"2006-01",    // year and month
		"2006",       // year only
	}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return &t
		}
	}
	return nil
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

func handleSlug(slug, title string) string {
	if slug == "" {
		slug = titleToSlug(title)
	}
	once := url.PathEscape(slug)
	twice := url.PathEscape(once)
	return twice
}
