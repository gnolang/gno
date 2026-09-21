package blog

import (
	"strings"
	"time"
)

func breadcrumb(parts []string) string {
	return "# " + strings.Join(parts, " / ") + "\n\n"
}

func getTitleKey(title string) string {
	return strings.ReplaceAll(title, " ", "")
}

func getPublishedKey(t time.Time) string {
	return t.Format(time.RFC3339)
}
