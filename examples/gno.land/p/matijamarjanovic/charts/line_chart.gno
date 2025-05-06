package charts

import (
	"strings"

	"gno.land/p/demo/ufmt"
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
	if pointSpacing < 5 || pointSpacing > 25 {
		return "pointSpacing must be between 5 and 25"
	}

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

	for i := 0; i < len(displayValues); i++ {
		x := i * (pointSpacing + 1)
		y := int(displayValues[i] * scale)
		if y > height {
			y = height
		}
		grid[height-y][x] = '*'

		if i < len(displayValues)-1 {
			nextX := (i + 1) * (pointSpacing + 1)
			nextY := int(displayValues[i+1] * scale)
			if nextY > height {
				nextY = height
			}
			dx := nextX - x
			dy := nextY - y
			prevInterpY := y
			for step := 1; step < dx; step++ {
				interpY := y + int(float64(dy)*float64(step)/float64(dx))
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
	if title != "" {
		output += "## " + title + "\n"
	}
	output += "```\n"

	if yAxisTitle != "" {
		output += yAxisTitle + "\n\n"
	}

	maxYValue := ufmt.Sprintf("%.2f", maxVal)
	yAxisWidth := len(maxYValue)

	for row := 0; row <= height; row++ {
		gridRow := grid[row]
		gridY := height - row
		if gridY%3 == 0 {
			yValue := ufmt.Sprintf("%.2f", float64(gridY)/scale)
			padding := strings.Repeat(" ", yAxisWidth-len(yValue))
			output += ufmt.Sprintf("%s%s | ", padding, yValue)
		} else {
			output += strings.Repeat(" ", yAxisWidth) + " | "
		}
		output += string(gridRow)
		output += "\n"
	}

	output += strings.Repeat(" ", yAxisWidth) + " +"
	output += strings.Repeat("-", gridWidth + 3)
	output += "\n"

	output += strings.Repeat(" ", yAxisWidth+2)
	for i := 0; i < len(displayValues); i++ {
		x := i * (pointSpacing + 1)
		valStr := ufmt.Sprintf("%.2f", displayValues[i])
		labelPos := x - len(valStr)/2
		if labelPos < 0 {
			labelPos = 0
		}
		if i == 0 {
			output += strings.Repeat(" ", labelPos)
		} else {
			prevX := (i - 1) * (pointSpacing + 1)
			prevValStr := ufmt.Sprintf("%.2f", displayValues[i-1])
			prevLabelEnd := prevX - len(prevValStr)/2 + len(prevValStr)
			output += strings.Repeat(" ", labelPos-prevLabelEnd)
		}
		output += valStr
	}
	output += "\n"

	if xAxisTitle != "" {
		padding := strings.Repeat(" ", gridWidth/2-len(xAxisTitle)/2)
		output += "       " + padding + xAxisTitle + "\n"
	}

	output += "```"
	return output
}
