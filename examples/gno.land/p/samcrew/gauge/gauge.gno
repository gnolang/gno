// Package gauge provides a simple way to render a gauge bar in SVG format.
package gauge

import (
	"gno.land/p/demo/svg"
	"gno.land/p/nt/ufmt"
)

// Render generates an SVG gauge bar.
// Parameters:
// - value: current value (must be <= total)
// - total: maximum value (must be > 0)
// - label: text label to display on the left side of the gauge
// - color: color of the filled part of the gauge (e.g., "#4caf50")
// - percentOnly: if true, only display the percentage on the right side; otherwise, display "value / total · percentage"
// - width: width of the gauge in pixels
func Render(value int, total int, label string, color string, percentOnly bool, width int) string {
	canvasWidth := width
	const (
		canvasHeight = 30
		fontSize     = 16
		padding      = 6
	)
	canvas := svg.NewCanvas(canvasWidth, canvasHeight)

	ratio := 0.0
	if total > 0 {
		ratio = float64(value) / float64(total)
	}

	leftText := label
	rightText := ufmt.Sprintf("%d / %d · %0.0f%%", value, total, ratio*100)
	if percentOnly {
		rightText = ufmt.Sprintf("%0.0f%%", ratio*100)
	}

	// Background
	canvas.Append(svg.Rectangle{
		X:      0,
		Y:      0,
		Width:  canvasWidth,
		Height: canvasHeight,
		Fill:   "#e0e0e0",
		RX:     2,
		RY:     2,
	})
	// Filled bar
	canvas.Append(svg.Rectangle{
		X:      0,
		Y:      0,
		Width:  int(ratio * float64(canvasWidth)),
		Height: canvasHeight,
		Fill:   color,
		RX:     2,
		RY:     2,
	})
	// Left text (label)
	canvas.Append(svg.Text{
		X:    padding,
		Y:    canvasHeight / 2,
		Text: leftText,
		Fill: "#000000",
		Attr: svg.BaseAttrs{
			Style: ufmt.Sprintf(
				"font-family:'Inter var',sans-serif;font-size:%dpx;dominant-baseline:middle;text-anchor:start;",
				fontSize,
			),
		},
	})
	// Right text (ratio + %)
	canvas.Append(svg.Text{
		X:    canvasWidth - padding,
		Y:    canvasHeight / 2,
		Text: rightText,
		Fill: "#000000",
		Attr: svg.BaseAttrs{
			Style: ufmt.Sprintf(
				"font-family:'Inter var',sans-serif;font-size:%dpx;dominant-baseline:middle;text-anchor:end;",
				fontSize,
			),
		},
	})

	return canvas.Render("Jauge")
}
