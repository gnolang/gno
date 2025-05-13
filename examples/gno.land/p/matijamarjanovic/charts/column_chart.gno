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

	maxVal := FindMaxValue(values)

	displayValues := values
	if len(values) > maxColumns {
		displayValues = normalizeData(values, maxColumns)
	}

	scale := height / maxVal

	output := FormatChartHeader(title)
	output += FormatYAxisTitle(yAxisTitle)

	maxYValue := ufmt.Sprintf("%.2f", maxVal)
	yAxisWidth := len(maxYValue)

	maxValStrLen := CalculateMaxValueStringLength(displayValues)
	if maxValStrLen < 2 {
		maxValStrLen = 2
	}
	colWidth := maxValStrLen + 2

	for row := height; row > 0; row-- {
		if row%3 == 0 {
			output += FormatYAxisLabel(float64(row), scale, yAxisWidth)
		} else {
			output += FormatYAxisSpace(yAxisWidth)
		}

		for _, val := range displayValues {
			scaledHeight := int(val * scale)
			if scaledHeight >= row {
				barCount := colWidth - 1
				output += strings.Repeat("|", barCount) + " "
			} else {
				output += strings.Repeat(" ", colWidth)
			}
		}
		output += "\n"
	}

	output += strings.Repeat(" ", yAxisWidth) + " +"
	output += strings.Repeat(strings.Repeat("-", colWidth), len(displayValues))
	output += "\n"

	output += FormatValueLabels(displayValues, colWidth, yAxisWidth)

	if xAxisTitle != "" {
		totalWidth := len(displayValues) * colWidth
		output += FormatXAxisTitle(xAxisTitle, totalWidth)
	}

	output += "```\n"
	return output
}
