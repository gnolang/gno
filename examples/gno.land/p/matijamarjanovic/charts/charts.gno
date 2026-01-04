// Charts is a package that provides functions to generate ASCII charts in markdown format.
// It includes functions for line charts, bar charts, and column charts.
//
// Both line and column charts are implemented in a way that offers normalization of data;
// In short terms precision of the chart can be adjusted by changing the number of points or columns.
// Line chart offers a way to adjust the spacing between points as well, which allows for more control over the chart.
//
// e.g. Let's say there is a data set of a value of dollar per day in the year of 2025 in a slice of floats.
// If those values are passed to the `GenerateLineChart` in this manner:
// `GenerateLineChart(values, 12, 5, "Dollar per day", "Days", "Dollars")`
// The chart will be normalized to 12 points, with 5 characters of spacing between each point, but more importantly,
// these points will represent a rough monthly average of the dollar value (months are not of equal length, so this is not an exact science).

package charts

import (
	"strings"

	"gno.land/p/nt/ufmt"
)

const height = 15

// normalizeData reduces the number of data points to maxColumns by averaging
// Returns nil if values is empty or maxColumns is zero or negative
func normalizeData(values []float64, maxColumns int) []float64 {
	if len(values) == 0 || maxColumns <= 0 {
		return nil
	}

	if maxColumns >= len(values) {
		result := make([]float64, len(values))
		copy(result, values)
		return result
	}

	result := make([]float64, maxColumns)
	groupSize := float64(len(values)) / float64(maxColumns)

	for i := 0; i < maxColumns; i++ {
		start := int(float64(i) * groupSize)
		end := int(float64(i+1) * groupSize)
		if end > len(values) {
			end = len(values)
		}

		sum := 0.0
		count := end - start
		for j := start; j < end; j++ {
			sum += values[j]
		}
		if count > 0 {
			result[i] = sum / float64(count)
		}
	}

	return result
}

// findMaxValue returns the maximum value in a slice of float64
func findMaxValue(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	maxVal := values[0]
	for _, v := range values[1:] {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}

// calculateMaxValueStringLength returns the maximum string length when formatting values
func calculateMaxValueStringLength(values []float64) int {
	maxValStrLen := 0
	for _, v := range values {
		valStr := ufmt.Sprintf("%.2f", v)
		if len(valStr) > maxValStrLen {
			maxValStrLen = len(valStr)
		}
	}
	return maxValStrLen
}

// formatChartHeader formats the chart header with title and markdown code block
func formatChartHeader(title string) string {
	output := ""
	if title != "" {
		output += "## " + title + "\n"
	}
	return output
}

// formatYAxisTitle adds the y-axis title to the chart
func formatYAxisTitle(yAxisTitle string) string {
	if yAxisTitle != "" {
		return yAxisTitle + "\n\n"
	}
	return ""
}

// formatXAxisTitle adds the x-axis title to the chart
func formatXAxisTitle(xAxisTitle string, width int) string {
	if xAxisTitle != "" {
		padding := strings.Repeat(" ", width/2-len(xAxisTitle)/2)
		return "       " + padding + xAxisTitle + "\n"
	}
	return ""
}

// formatYAxisLabel formats a y-axis label with proper padding
func formatYAxisLabel(value float64, scale float64, yAxisWidth int) string {
	yValue := ufmt.Sprintf("%.2f", value/scale)
	padding := strings.Repeat(" ", yAxisWidth-len(yValue))
	return ufmt.Sprintf("%s%s | ", padding, yValue)
}

// formatYAxisSpace returns spacing for y-axis when no label is needed
func formatYAxisSpace(yAxisWidth int) string {
	return strings.Repeat(" ", yAxisWidth) + " | "
}
