package charts

import (
	"strings"

	"gno.land/p/demo/svg"
	"gno.land/p/nt/ufmt"
)

// GenerateBarChartMarkdown creates an ASCII bar chart in markdown format
// values: slice of float values to chart
// labels: slice of labels for each bar
// maxWidth: maximum width of bars in characters
// title: chart title
// Returns a markdown string representing the chart
func GenerateBarChartMarkdown(values []float64, labels []string, maxWidth int, title string) string {
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

// GenerateBarChart renders a bar chart as SVG.
// width/height include margins.
func GenerateBarChart(values []float64, labels []string, width, height int) string {
	if len(values) == 0 || len(labels) == 0 || len(values) != len(labels) {
		return "invalid data for display"
	}

	if width <= 0 || height <= 0 {
		return "canvas size must be > 0"
	}

	maxVal := findMaxValue(values)
	if maxVal <= 0 {
		maxVal = 1
	}

	left, right, top, bottom := 60.0, 20.0, 30.0, 50.0
	plotW := float64(width) - left - right
	plotH := float64(height) - top - bottom

	barCount := len(values)
	gapRatio := 0.3 // 30% of slot is gap
	slotW := plotW / float64(barCount)
	barW := slotW * (1.0 - gapRatio)
	gapW := slotW - barW

	c := svg.NewCanvas(width, height)

	c.Append(svg.Line{X1: int(left), Y1: int(top + plotH), X2: int(left + plotW), Y2: int(top + plotH), Stroke: "#CBD5E1"})
	c.Append(svg.Line{X1: int(left), Y1: int(top), X2: int(left), Y2: int(top + plotH), Stroke: "#CBD5E1"})

	ticks := 5
	for i := 0; i <= ticks; i++ {
		y := top + plotH*(float64(i)/float64(ticks))
		c.Append(svg.Line{X1: int(left), Y1: int(y), X2: int(left + plotW), Y2: int(y), Stroke: "#E2E8F0", Attr: svg.BaseAttrs{Opacity: "0.4"}})
		val := maxVal * (1.0 - float64(i)/float64(ticks))
		c.Append(svg.Text{X: int(left) - 8, Y: int(y) + 5, Text: ufmt.Sprintf("%.2f", val), Fill: "#64748B", Attr: svg.BaseAttrs{Style: "text-anchor:end; font-size:12px; font-family:'Inter var',sans-serif;"}})
	}

	for i, v := range values {
		hpx := 0.0
		if maxVal > 0 {
			hpx = (v / maxVal) * plotH
		}
		x := left + float64(i)*slotW + gapW/2
		y := top + (plotH - hpx)
		c.Append(svg.Rectangle{X: int(x), Y: int(y), Width: int(barW), Height: int(hpx), Fill: "#6366F1"})

		c.Append(svg.Text{X: int(x + barW/2), Y: int(top + plotH + 18), Text: labels[i], Fill: "#475569", Attr: svg.BaseAttrs{Style: "text-anchor:middle; font-size:12px; font-family:'Inter var',sans-serif;"}})
	}

	return c.Render("Bar Chart")
}
