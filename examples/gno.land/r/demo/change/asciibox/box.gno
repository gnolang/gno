package asciibox

import (
	"strings"
)

func getLongest(table []string) string {
	maxLen := -1
	res := ""
	for _, e := range table {
		if len(e) <= maxLen {
			continue
		}
		maxLen = len(e)
		res = e
	}
	return res
}

func BoxThat(name string, table []string, size int) string {
	char := []string{"─", "│", "┌", "┐", "└", "┘"}
	longest := getLongest(table)

	longLen := len(longest)
	if size > longLen {
		longLen = size
	}
	if longLen <= 0 {
		longLen = len(name) + 4
	}

	var box strings.Builder

	title := ""
	if name != "" {
		title = " " + name + " "
	} else {
		title = char[0]
	}

	end := strings.Repeat(char[0], len(title))
	line := strings.Repeat(char[0], max(0, longLen-(len(title)-2)))

	box.WriteString("```\n")
	box.WriteString(char[2] + title + line + char[3] + "\n")

	for _, text := range table {
		padding := max(longLen, len(title)-2) - len(text)
		padded := text + strings.Repeat(" ", padding)
		box.WriteString(char[1] + " " + padded + " " + char[1] + "\n")
	}

	box.WriteString(char[4] + end + line + char[5] + "\n")
	box.WriteString("```\n")

	return box.String()
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
