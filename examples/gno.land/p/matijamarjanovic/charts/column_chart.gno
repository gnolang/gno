package charts

import (
	"strings"

	"gno.land/p/demo/ufmt"
)

// GenerateColumnChart creates an ASCII column chart in markdown format
// values: slice of float values to chart
// maxColumns: maximum number of columns to display (will normalize if exceeded)
// title: chart title
// xAxisTitle: title for the x-axis
// yAxisTitle: title for the y-axis
// Returns a markdown string representing the chart
func GenerateColumnChart(values []float64, maxColumns int, title, xAxisTitle, yAxisTitle string) string {
	if len(values) == 0 {
		return "no data to display"
	}

	maxVal := values[0]
	for _, v := range values[1:] {
		if v > maxVal {
			maxVal = v
		}
	}

	displayValues := values
	if len(values) > maxColumns {
		displayValues = normalizeData(values, maxColumns)
	}

	scale := height / maxVal

	output := ""
	if title != "" {
		output += "## " + title + "\n"
	}

	output += "\n```\n"

	if yAxisTitle != "" {
		output += yAxisTitle + "\n\n"
	}

	maxYValue := ufmt.Sprintf("%.2f", maxVal)
	yAxisWidth := len(maxYValue)

	for row := height; row > 0; row-- {
		if row%3 == 0 {
			yValue := ufmt.Sprintf("%.2f", float64(row)/scale)
			padding := strings.Repeat(" ", yAxisWidth-len(yValue))
			output += ufmt.Sprintf("%s%s | ", padding, yValue)
		} else {
			output += strings.Repeat(" ", yAxisWidth) + " | "
		}

		for _, val := range displayValues {
			scaledHeight := int(val * scale)
			if scaledHeight >= row {
				output += "||||| "
			} else {
				output += "      "
			}
		}
		output += "\n"
	}

	output += strings.Repeat(" ", yAxisWidth) + " +"
	output += strings.Repeat("------", len(displayValues))
	output += strings.Repeat(" ", len(displayValues))
	output += "\n"

	output += strings.Repeat(" ", yAxisWidth+2) + " "
	for _, val := range displayValues {
		output += ufmt.Sprintf("%.2f  ", val)
	}
	output += "\n"

	if xAxisTitle != "" {
		padding := strings.Repeat(" ", (len(displayValues)*6)/2)
		output += "       " + padding + xAxisTitle + "\n"
	}

	output += "```\n"
	return output
}
