package gnoffee

import (
	"regexp"
)

// Stage1 converts the gnoffee-specific keywords into their comment directive equivalents.
func Stage1(src string) string {
	// Handling the 'export' keyword
	exportRegex := regexp.MustCompile(`(?m)^export\s+`)
	src = exportRegex.ReplaceAllString(src, "//gnoffee:export ")

	// Handling the 'invar' keyword
	invarRegex := regexp.MustCompile(`(?m)^invar\s+([\w\d_]+)\s+(.+)`)
	src = invarRegex.ReplaceAllString(src, "//gnoffee:invar $1\nvar $1 $2")

	return src
}
