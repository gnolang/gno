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

	maxValStrLen := 0
	for _, v := range displayValues {
		valStr := ufmt.Sprintf("%.2f", v)
		if len(valStr) > maxValStrLen {
			maxValStrLen = len(valStr)
		}
	}
	if maxValStrLen < 2 {
		maxValStrLen = 2
	}
	colWidth := maxValStrLen + 2

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

	output += strings.Repeat(" ", yAxisWidth+2)
	for i := 0; i < len(displayValues); i++ {
		x := i * colWidth
		valStr := ufmt.Sprintf("%.2f", displayValues[i])
		labelPos := x + colWidth/2 - len(valStr)/2
		if labelPos < 0 {
			labelPos = 0
		}
		if i == 0 {
			output += strings.Repeat(" ", labelPos)
		} else {
			prevX := (i - 1) * colWidth
			prevValStr := ufmt.Sprintf("%.2f", displayValues[i-1])
			prevLabelEnd := prevX + colWidth/2 - len(prevValStr)/2 + len(prevValStr)
			output += strings.Repeat(" ", labelPos-prevLabelEnd)
		}
		output += valStr
	}
	output += "\n"

	if xAxisTitle != "" {
		padding := strings.Repeat(" ", (len(displayValues)*colWidth)/2)
		output += "       " + padding + xAxisTitle + "\n"
	}

	output += "```\n"
	return output
}
