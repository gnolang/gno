package markdown_utils

import (
	"strings"
)

func Indent(markdown string) string {
	lines := strings.Split(markdown, "\n")
	for i, line := range lines {
		lines[i] = "  " + line
	}
	return strings.Join(lines, "\n")
}
