// Package gauge provides a simple way to render a gauge bar in SVG format.
package gauge

import (
	"gno.land/p/demo/svg"
	"gno.land/p/demo/ufmt"
)

// TODO: Doc
func Render(value int, total int, label string, color string, percentOnly bool, width int) string {
	if total <= 0 {
		return "jauge fails: total must be greater than 0"
	}
	if value > total {
		return "jauge fails: value cannot be greater than total"
	}

	canvasWidth := width
	const (
		canvasHeight = 30
		fontSize     = 16
		padding      = 6
	)

	println("total", total)
	println("value", value)

	ratio := 0.0
	if total > 0 {
		ratio = float64(value) / float64(total)
		println("value / total", value/total)

	}

	println("ratio", ratio)

	leftText := label
	rightText := ufmt.Sprintf("%d / %d · %0.0f%%", value, total, ratio*100)
	if percentOnly {
		rightText = ufmt.Sprintf("%0.0f%%", ratio*100)
	}

	canvas := svg.NewCanvas(canvasWidth, canvasHeight)

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
