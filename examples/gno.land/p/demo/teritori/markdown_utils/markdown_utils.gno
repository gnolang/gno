package markdown_utils

import (
	"strings"
)

// this function take as input a markdown string and add an indentation level to markdown titles
func Indent(markdown string) string {
	// split the markdown string into lines
	lines := strings.Split(markdown, "\n")

	// iterate over the lines
	for i, line := range lines {
		// if the line starts with a markdown title
		if strings.HasPrefix(line, "#") {
			// add an indentation level to the title
			lines[i] = "#" + line
		}
	}

	// join the lines back into a string
	return strings.Join(lines, "\n")
}

// thanks copilot this is perfect xD
// I just renamed it, AddIndentationLevelToMarkdownTitles was too long
