package svgbtn

import (
	"gno.land/p/demo/svg"
	"gno.land/p/demo/ufmt"
)

func Button(width, height int, btnColor, textColor, text, link string) string {
	// Create canvas
	canvas := svg.NewCanvas(width, height).
		WithViewBox(0, 0, width, height).
		AddStyle("text", "font-family:sans-serif;font-size:14px;text-anchor:middle;dominant-baseline:middle;")

	// color given in HEX WITH #
	bg := svg.NewRectangle(0, 0, width, height, btnColor)

	// Create label
	label := svg.NewText(width/2, height/2, text, textColor)

	// Add elements to canvas
	canvas.Append(bg, label)

	// Return the base64 markdown image string
	return ufmt.Sprintf("[%s](%s) ", canvas.Render(text), link)
}
