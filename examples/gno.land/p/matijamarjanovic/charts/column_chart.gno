package charts

import (
	"strings"

	"gno.land/p/demo/svg"
	"gno.land/p/nt/ufmt"
)

// GenerateColumnChartMarkdown creates an ASCII column chart in markdown format
// values: slice of float values to chart
// maxColumns: maximum number of columns to display (will normalize if exceeded)
// title: chart title
// xAxisTitle: title for the x-axis
// yAxisTitle: title for the y-axis
// Returns a markdown string representing the chart
func GenerateColumnChartMarkdown(values []float64, maxColumns int, title, xAxisTitle, yAxisTitle string) string {
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

// GenerateColumnChart renders a normalized column chart as SVG.
func GenerateColumnChart(values []float64, maxColumns int, width, height int, xAxisTitle, yAxisTitle string) string {
	if len(values) == 0 {
		return "no data to display"
	}
	if maxColumns <= 0 || width <= 0 || height <= 0 {
		return "invalid params"
	}

	displayValues := values
	if len(values) > maxColumns {
		displayValues = normalizeData(values, maxColumns)
	}

	maxVal := findMaxValue(displayValues)
	if maxVal <= 0 {
		maxVal = 1
	}

	left, right, top, bottom := 50.0, 20.0, 30.0, 50.0
	plotW := float64(width) - left - right
	plotH := float64(height) - top - bottom

	colCount := len(displayValues)
	gapRatio := 0.4
	slotW := plotW / float64(colCount)
	colW := slotW * (1.0 - gapRatio)
	gapW := slotW - colW

	c := svg.NewCanvas(width, height)

	c.Append(svg.Line{X1: int(left), Y1: int(top + plotH), X2: int(left + plotW), Y2: int(top + plotH), Stroke: "#CBD5E1"})
	c.Append(svg.Line{X1: int(left), Y1: int(top), X2: int(left), Y2: int(top + plotH), Stroke: "#CBD5E1"})

	ticks := 5
	for i := 0; i <= ticks; i++ {
		y := top + plotH*(float64(i)/float64(ticks))
		c.Append(svg.Line{X1: int(left), Y1: int(y), X2: int(left + plotW), Y2: int(y), Stroke: "#F1F5F9"})
	}

	for i, v := range displayValues {
		hpx := (v / maxVal) * plotH
		x := left + float64(i)*slotW + gapW/2
		y := top + (plotH - hpx)
		c.Append(svg.Rectangle{X: int(x), Y: int(y), Width: int(colW), Height: int(hpx), Fill: "#22C55E"})
	}

	for i := range displayValues {
		x := left + float64(i)*slotW + slotW/2
		c.Append(svg.Text{X: int(x), Y: int(top + plotH + 18), Text: ufmt.Sprintf("%d", i+1), Fill: "#475569", Attr: svg.BaseAttrs{Style: "text-anchor:middle; font-size:12px; font-family:'Inter var',sans-serif;"}})
	}

	if yAxisTitle != "" {
		c.Append(svg.Text{X: int(left) - 35, Y: int(top) - 10, Text: yAxisTitle, Fill: "#334155", Attr: svg.BaseAttrs{Style: "font-size:12px; font-family:'Inter var',sans-serif;"}})
	}
	if xAxisTitle != "" {
		c.Append(svg.Text{X: int(left + plotW/2), Y: int(top + plotH + 38), Text: xAxisTitle, Fill: "#334155", Attr: svg.BaseAttrs{Style: "text-anchor:middle; font-size:12px; font-family:'Inter var',sans-serif;"}})
	}

	return c.Render("Column Chart")
}
