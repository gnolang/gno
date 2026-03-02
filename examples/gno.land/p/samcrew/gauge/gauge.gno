// Package gauge provides a simple way to render a gauge bar in SVG format.
package gauge

import (
	"gno.land/p/demo/svg"
	"gno.land/p/nt/ufmt/v0"
)

type Config struct {
	PercentOnly  bool
	Width        int
	CanvasHeight int
	FontSize     int
	PaddingH     int
}

var DefaultConfig = Config{
	PercentOnly:  false,
	Width:        300,
	CanvasHeight: 30,
	FontSize:     16,
	PaddingH:     6,
}

// Render generates an SVG gauge bar.
// Parameters:
// - value: current value (must be <= total)
// - total: maximum value (must be > 0)
// - label: text label to display on the left side of the gauge
// - color: color of the filled part of the gauge (e.g., "#4caf50")
// - config: configuration options:
//   - PercentOnly: Only display the percentage on the right side; otherwise, display "value / total · percentage"
//   - Width: Width of the gauge in pixels
//   - CanvasHeight: Height of the gauge in pixels
//   - FontSize: Font size of the text in pixels
//   - PaddingH: Horizontal padding (for the text) in pixels
//
// Example usage: gauge.Render(30, 100, "Progress", "#4caf50", &gauge.Config{PercentOnly: true, Width: 400})
func Render(value int, total int, label string, color string, config Config) string {
	if total <= 0 {
		return "jauge fails: total must be greater than 0"
	}
	if value <= 0 {
		return "jauge fails: value must be greater than 0"
	}
	if value > total {
		return "jauge fails: value cannot be greater than total"
	}

	canvasWidth := config.Width
	canvasHeight := config.CanvasHeight
	fontSize := config.FontSize
	paddingH := config.PaddingH
	canvas := svg.NewCanvas(canvasWidth, canvasHeight)

	ratio := 0.0
	if total > 0 {
		ratio = float64(value) / float64(total)
	}

	leftText := label
	rightText := ufmt.Sprintf("%d / %d · %0.0f%%", value, total, ratio*100)
	if config.PercentOnly {
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
		X:    paddingH,
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
		X:    canvasWidth - paddingH,
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
