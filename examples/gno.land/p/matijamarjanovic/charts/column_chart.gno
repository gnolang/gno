package charts

import (
	"strings"

	"gno.land/p/nt/ufmt"
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

	if maxColumns <= 0 {
		return "maxColumns must be greater than 0"
	}

	maxVal := findMaxValue(values)

	displayValues := values
	if len(values) > maxColumns {
		displayValues = normalizeData(values, maxColumns)
	}

	scale := height / maxVal

	output := formatChartHeader(title)
	output += "\n```\n"
	output += formatYAxisTitle(yAxisTitle)

	maxYValue := ufmt.Sprintf("%.2f", maxVal)
	yAxisWidth := len(maxYValue)

	maxValStrLen := calculateMaxValueStringLength(displayValues)
	if maxValStrLen < 2 {
		maxValStrLen = 2
	}
	colWidth := maxValStrLen + 2

	for row := height; row > 0; row-- {
		if row%3 == 0 {
			output += formatYAxisLabel(float64(row), scale, yAxisWidth)
		} else {
			output += formatYAxisSpace(yAxisWidth)
		}

		for _, val := range displayValues {
			scaledHeight := int(val * scale)
			if scaledHeight >= row {
				barWidth := colWidth - 1
				output += strings.Repeat("|", barWidth) + " "
			} else {
				output += strings.Repeat(" ", colWidth)
			}
		}
		output += "\n"
	}

	output += strings.Repeat(" ", yAxisWidth) + " +"
	output += strings.Repeat(strings.Repeat("-", colWidth), len(displayValues))
	output += "\n"

	output += formatValueLabelsColumnChart(displayValues, colWidth, yAxisWidth)

	if xAxisTitle != "" {
		totalWidth := len(displayValues) * colWidth
		output += formatXAxisTitle(xAxisTitle, totalWidth)
	}

	output += "```\n"
	return output
}

// formatValueLabelsColumnChart formats the value labels for the x-axis for column charts
func formatValueLabelsColumnChart(values []float64, colWidth int, yAxisWidth int) string {
	output := strings.Repeat(" ", yAxisWidth+2)

	for i := 0; i < len(values); i++ {
		x := i * colWidth
		valStr := ufmt.Sprintf("%.2f", values[i])
		labelPos := x + colWidth/2 - len(valStr)/2
		if labelPos < 0 {
			labelPos = 0
		}

		if i == 0 {
			output += strings.Repeat(" ", labelPos)
		} else {
			prevX := (i - 1) * colWidth
			prevValStr := ufmt.Sprintf("%.2f", values[i-1])
			prevLabelEnd := prevX + colWidth/2 - len(prevValStr)/2 + len(prevValStr)
			output += strings.Repeat(" ", labelPos-prevLabelEnd)
		}
		output += valStr
	}

	return output + "\n"
}
