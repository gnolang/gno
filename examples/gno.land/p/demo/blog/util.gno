package blog

import "strings"

func breadcrumb(parts []string) string {
	return "# " + strings.Join(parts, " / ") + "\n\n"
}
