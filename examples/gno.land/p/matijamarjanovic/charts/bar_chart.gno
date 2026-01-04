package charts

import (
	"strings"

	"gno.land/p/nt/ufmt"
)

// GenerateBarChart creates an ASCII bar chart in markdown format
// values: slice of float values to chart
// labels: slice of labels for each bar
// maxWidth: maximum width of bars in characters
// title: chart title
// Returns a markdown string representing the chart
func GenerateBarChart(values []float64, labels []string, maxWidth int, title string) string {
	if len(values) == 0 || len(labels) == 0 || len(values) != len(labels) {
		return "invalid data for display"
	}

	if maxWidth <= 0 {
		return "maxWidth must be greater than 0"
	}

	maxVal := findMaxValue(values)

	maxLabelLength := 0
	for _, label := range labels {
		if len(label) > maxLabelLength {
			maxLabelLength = len(label)
		}
	}

	scale := float64(maxWidth) / maxVal

	output := formatChartHeader(title)
	output += "\n```\n"

	for i, value := range values {
		padding := strings.Repeat(" ", maxLabelLength-len(labels[i]))
		output += labels[i] + padding + "    "

		barLength := int(value * scale)
		output += strings.Repeat("█", barLength)

		output += " " + ufmt.Sprintf("%.2f", value)
		output += "\n"
	}

	output += "```\n"
	return output
}
