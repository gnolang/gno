package charts

import (
	"gno.land/p/demo/ufmt"
	"strings"
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
		return "No data to display"
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

	const height = 15
	scale := height / maxVal

	output := ""
	if title != "" {
		output += "## " + title + "\n"
	}
	
	output += "```\n"

	for row := height; row > 0; row-- {
		if row%3 == 0 {
			yValue := ufmt.Sprintf("%f", float64(row)/scale)
			output += ufmt.Sprintf("%s |", yValue)
		} else {
			output += "      |"
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

	output += "      +"
	output += strings.Repeat("-----", len(displayValues))
	output += strings.Repeat(" ", len(displayValues))
	output += "\n"

	output += "       "
	for _, val := range displayValues {
		output += ufmt.Sprintf("%f  ", val)
	}
	output += "\n"

	if xAxisTitle != "" {
		padding := strings.Repeat(" ", (len(displayValues)*6)/2)
		output += "       " + padding + xAxisTitle + "\n"
	}

	if yAxisTitle != "" {
		lines := strings.Split(output, "\n")
		var newLines []string
		yAxisPos := (len(lines) - 4) / 2
		for i, line := range lines {
			if i == yAxisPos {
				newLines = append(newLines, yAxisTitle + " " + line)
			} else {
				newLines = append(newLines, line)
			}
		}
		output = strings.Join(newLines, "\n")
	}

	output += "```"
	return output
}

// normalizeData reduces the number of data points to maxColumns by averaging
func normalizeData(values []float64, maxColumns int) []float64 {
	result := make([]float64, maxColumns)
	groupSize := float64(len(values)) / float64(maxColumns)
	
	for i := 0; i < maxColumns; i++ {
		start := int(float64(i) * groupSize)
		end := int(float64(i+1) * groupSize)
		if end > len(values) {
			end = len(values)
		}
			
		sum := 0.0
		count := 0
		for j := start; j < end; j++ {
			sum += values[j]
			count++
		}
		result[i] = sum / float64(count)
	}
	
	return result
}
