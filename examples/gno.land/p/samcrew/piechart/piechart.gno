package piechart

import (
	"math"
	"strings"

	"gno.land/p/demo/svg"
	"gno.land/p/demo/ufmt"
	"gno.land/p/sunspirit/md"
)

// RenderPieChart creates an SVG pie chart from given values, colors, labels, and optional title.
// It returns the SVG markup as a string, including a markdown header if a title is provided.
func RenderPieChart(values []float64, colors []string, labels []string, title string) string {
	// Validate input slices length
	if len(values) != len(colors) || len(values) != len(labels) {
		return "\npiechart fails: values, colors and labels slices must have the same length"
	}

	const (
		canvasWidth  = 500
		canvasHeight = 200
		centerX      = 100.0
		centerY      = 100.0
		radius       = 80.0
		legendX      = 210
		legendStartY = 30
		lineHeight   = 26
		squareSize   = 16
		fontSize     = 16
	)

	canvas := svg.NewCanvas(canvasWidth, canvasHeight)

	// Sum all values to compute slice proportions
	var total float64
	for _, v := range values {
		total += v
	}

	startAngle := 0.0
	for i, v := range values {
		// Calculate angle span for current slice
		angle := (v / total) * 2 * math.Pi
		endAngle := startAngle + angle

		// Compute start and end points on the circle circumference
		x1 := centerX + radius*math.Cos(startAngle)
		y1 := centerY + radius*math.Sin(startAngle)
		x2 := centerX + radius*math.Cos(endAngle)
		y2 := centerY + radius*math.Sin(endAngle)

		// Determine if the arc should be a large arc (> 180 degrees) (Arc direction)
		largeArcFlag := 0
		if angle > math.Pi {
			largeArcFlag = 1
		}

		// Build the SVG path for the pie slice
		path := ufmt.Sprintf(
			"M%.2f,%.2f L%.2f,%.2f A%.2f,%.2f 0 %d 1 %.2f,%.2f Z",
			centerX, centerY, x1, y1, radius, radius, largeArcFlag, x2, y2,
		)

		canvas.Append(svg.Path{
			D:    path,
			Fill: colors[i],
		})

		startAngle = endAngle
	}

	// Draw legend with colored squares and labels including percentages
	for i, label := range labels {
		y := legendStartY + i*lineHeight

		// Colored square representing slice color
		canvas.Append(svg.Rectangle{
			X:      legendX,
			Y:      y - squareSize/2,
			Width:  squareSize,
			Height: squareSize,
			Fill:   colors[i],
		})

		// Legend text showing label, value and percentage
		text := ufmt.Sprintf("%s: %.0f (%.1f%%)", label, values[i], values[i]*100/total)
		canvas.Append(svg.Text{
			X:    legendX + squareSize + 8,
			Y:    y + fontSize/3,
			Text: text,
			Fill: "#54595D",
			Attr: svg.BaseAttrs{
				Style: ufmt.Sprintf("font-family:'Inter var',sans-serif;font-size:%dpx;", fontSize),
			},
		})
	}

	return formatChartHeader(title) + canvas.Render("Pie Chart "+title)
}

// formatChartHeader returns a markdown header for the chart title if provided.
func formatChartHeader(title string) string {
	var sb strings.Builder
	if title == "" {
		return sb.String()
	}
	sb.WriteString(md.H2(title))
	return sb.String()
}
