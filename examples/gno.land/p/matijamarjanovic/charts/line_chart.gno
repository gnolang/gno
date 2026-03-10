package charts

import (
	"strings"

	"gno.land/p/nt/ufmt/v0"
)

// GenerateLineChart creates an ASCII line chart in markdown format
// values: slice of float values to chart
// maxPoints: maximum number of points to display (will normalize if exceeded)
// pointSpacing: horizontal spacing between points (in characters)
// title: chart title
// xAxisTitle: title for the x-axis
// yAxisTitle: title for the y-axis
// Returns a markdown string representing the chart
func GenerateLineChart(values []float64, maxPoints int, pointSpacing int, title, xAxisTitle, yAxisTitle string) string {
	if len(values) == 0 {
		return "no data to display"
	}

	if maxPoints <= 0 {
		return "maxPoints must be greater than 0"
	}

	maxVal := findMaxValue(values)
	maxValStrLen := calculateMaxValueStringLength(values)

	minSpacing := maxValStrLen + 2
	if pointSpacing < minSpacing {
		return ufmt.Sprintf("pointSpacing must be at least %d to fit value labels", minSpacing)
	}
	if pointSpacing > 25 {
		return ufmt.Sprintf("pointSpacing must be between %d and 25", minSpacing)
	}

	displayValues := values
	if len(values) > maxPoints {
		displayValues = normalizeData(values, maxPoints)
	}

	scale := float64(height) / maxVal

	gridWidth := (len(displayValues)-1)*pointSpacing + len(displayValues)
	grid := make([][]rune, height+1)
	for i := range grid {
		grid[i] = make([]rune, gridWidth)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	spacing := pointSpacing + 1
	for i := 0; i < len(displayValues); i++ {
		x := i * spacing
		y := int(displayValues[i] * scale)
		if y > height {
			y = height
		}
		grid[height-y][x] = '*'

		if i < len(displayValues)-1 {
			nextX := (i + 1) * spacing
			nextY := int(displayValues[i+1] * scale)
			if nextY > height {
				nextY = height
			}
			dx := nextX - x
			dy := nextY - y
			prevInterpY := y
			for step := 1; step < dx; step++ {
				interpY := y + (dy*step)/dx
				startY := prevInterpY
				endY := interpY
				if startY > endY {
					startY, endY = endY, startY
				}
				for fillY := startY; fillY <= endY; fillY++ {
					grid[height-fillY][x+step] = '*'
				}
				prevInterpY = interpY
			}
		}
	}

	output := "\n"
	output += formatChartHeader(title)
	output += "\n```\n"
	output += formatYAxisTitle(yAxisTitle)

	maxYValue := ufmt.Sprintf("%.2f", maxVal)
	yAxisWidth := len(maxYValue)

	for row := 0; row <= height; row++ {
		gridRow := grid[row]
		gridY := height - row
		if gridY%3 == 0 {
			output += formatYAxisLabel(float64(gridY), scale, yAxisWidth)
		} else {
			output += formatYAxisSpace(yAxisWidth)
		}
		output += string(gridRow)
		output += "\n"
	}

	output += strings.Repeat(" ", yAxisWidth) + " +"
	output += strings.Repeat("-", gridWidth+3)
	output += "\n"

	output += formatValueLabelsLineChart(displayValues, pointSpacing, yAxisWidth)

	if xAxisTitle != "" {
		output += formatXAxisTitle(xAxisTitle, gridWidth)
	}

	output += "```"
	return output
}

// formatValueLabelsLineChart formats the value labels for the x-axis for line charts
func formatValueLabelsLineChart(displayValues []float64, pointSpacing int, yAxisWidth int) string {
	output := strings.Repeat(" ", yAxisWidth+2)
	spacing := pointSpacing + 1
	for i := 0; i < len(displayValues); i++ {
		x := i * spacing
		valStr := ufmt.Sprintf("%.2f", displayValues[i])
		labelPos := x - len(valStr)/2
		if labelPos < 0 {
			labelPos = 0
		}
		if i == 0 {
			output += strings.Repeat(" ", labelPos)
		} else {
			prevX := (i - 1) * spacing
			prevValStr := ufmt.Sprintf("%.2f", displayValues[i-1])
			prevLabelEnd := prevX - len(prevValStr)/2 + len(prevValStr)
			output += strings.Repeat(" ", labelPos-prevLabelEnd)
		}
		output += valStr
	}
	return output + "\n"
}
