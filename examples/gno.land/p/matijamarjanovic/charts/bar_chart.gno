package charts

import (
	"gno.land/p/demo/ufmt"
	"strings"
)

// GenerateBarChart creates an ASCII bar chart in markdown format
// values: slice of float values to chart
// labels: slice of labels for each bar
// maxWidth: maximum width of bars in characters
// title: chart title
// Returns a markdown string representing the chart
func GenerateBarChart(values []float64, labels []string, maxWidth int, title string) string {
	if len(values) == 0 || len(labels) == 0 || len(values) != len(labels) {
		return "Invalid data for display"
	}

	maxVal := values[0]
	maxLabelLength := len(labels[0])
	for i, v := range values {
		if v > maxVal {
			maxVal = v
		}
		if len(labels[i]) > maxLabelLength {
			maxLabelLength = len(labels[i])
		}
	}

	scale := float64(maxWidth) / maxVal

	output := ""
	if title != "" {
		output += "## " + title + "\n"
	}
	output += "```\n"

	for i, value := range values {
		padding := strings.Repeat(" ", maxLabelLength-len(labels[i]))
		output += labels[i] + padding + "    "

		barLength := int(value * scale)
		output += strings.Repeat("█", barLength)

		output += " " + ufmt.Sprintf("%.0f", value)
		output += "\n"
	}

	output += "```"
	return output
} 