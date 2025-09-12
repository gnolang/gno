// Package piechart provides functionality to render a pie chart as an SVG image.
// It takes a list of PieSlice objects, each representing a slice of the pie with a value,
// color, and label, and generates an SVG representation of the pie chart.
package piechart

import (
	"math"

	"gno.land/p/demo/svg"
	"gno.land/p/nt/ufmt"
	"gno.land/p/sunspirit/md"
)

type PieSlice struct {
	Value float64
	Color string
	Label string
}

// Render creates an SVG pie chart from given slices (value, color and label).
// It returns an img svg markup as a string, including a markdown header if a non-empty title is provided.
func Render(slices []PieSlice, title string) string {
	// Validate input slices length
	if len(slices) == 0 {
		return "\npiechart fails: no data provided"
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

	// Sum all values to compute slices proportions
	var total float64
	for _, s := range slices {
		total += s.Value
	}

	// Draw pie slices and legend in one pass
	startAngle := -math.Pi / 2
	for i, s := range slices {
		if s.Value > 0 {
			// --- PIE SLICE ---
			// Calculate angle span for current slice
			angle := (s.Value / total) * 2 * math.Pi
			endAngle := startAngle + angle

			// Compute start and end points on the circle circumference
			cosStart, sinStart := math.Cos(startAngle), math.Sin(startAngle)
			cosEnd, sinEnd := math.Cos(endAngle), math.Sin(endAngle)
			x1 := centerX + radius*cosStart
			y1 := centerY + radius*sinStart
			x2 := centerX + radius*cosEnd
			y2 := centerY + radius*sinEnd

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

			// Colored slice
			canvas.Append(svg.Path{
				D:    path,
				Fill: s.Color,
			})

			startAngle = endAngle
		}

		// --- LEGEND ---
		y := legendStartY + i*lineHeight
		// Colored square representing slice color
		canvas.Append(svg.Rectangle{
			X:      legendX,
			Y:      y - squareSize/2,
			Width:  squareSize,
			Height: squareSize,
			Fill:   s.Color,
		})

		// Legend text showing label, value and percentage
		text := ufmt.Sprintf("%s: %.0f (%.1f%%)", s.Label, s.Value, s.Value*100/total)
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

	if title == "" {
		return canvas.Render("Pie Chart")
	}
	return md.H2(title) + canvas.Render("Pie Chart "+title)
}
