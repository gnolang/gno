package svgbtn

import (
	"gno.land/p/demo/svg"
	"gno.land/p/demo/ufmt"
)

// Button creates a base SVG button with given size, colors, label, and link.
// - `width`, `height`: size in pixels
// - `btnColor`: background color (e.g. "#007BFF")
// - `textColor`: label color (e.g. "#FFFFFF")
// - `text`: visible button label
// - `link`: URL to wrap the image in markdown-style [svg](link)
func Button(width, height int, btnColor, textColor, text, link string) string {
	canvas := svg.NewCanvas(width, height).
		WithViewBox(0, 0, width, height).
		AddStyle("text", "font-family:sans-serif;font-size:14px;text-anchor:middle;dominant-baseline:middle;")

	bg := svg.NewRectangle(0, 0, width, height, btnColor)
	bg.RX = height / 5
	bg.RY = height / 5

	label := svg.NewText(width/2, height/2, text, textColor)

	canvas.Append(bg, label)

	return ufmt.Sprintf("[%s](%s)", canvas.Render(text), link)
}

// DisabledButton renders a button without a link, grayed out to indicate disabled state.
func DisabledButton(width, height int, text string) string {
	canvas := svg.NewCanvas(width, height).
		WithViewBox(0, 0, width, height)

	bg := svg.NewRectangle(0, 0, width, height, "#cccccc")
	bg.RX = height / 5
	bg.RY = height / 5

	label := svg.NewText(width/2, height/2, text, "#666666")

	canvas.Append(bg, label)

	return canvas.Render(text)
}

// PrimaryButton renders a blue button with white text.
func PrimaryButton(width, height int, text, link string) string {
	return Button(width, height, "#007BFF", "#ffffff", text, link)
}

// DangerButton renders a red button with white text.
func DangerButton(width, height int, text, link string) string {
	return Button(width, height, "#DC3545", "#ffffff", text, link)
}

// SuccessButton renders a green button with white text.
func SuccessButton(width, height int, text, link string) string {
	return Button(width, height, "#28A745", "#ffffff", text, link)
}

// SmallButton renders a compact gray button with white text.
func SmallButton(width, height int, text, link string) string {
	return Button(width, height, "#6C757D", "#ffffff", text, link)
}

// WideButton renders a wider cyan button with white text.
func WideButton(width, height int, text, link string) string {
	return Button(width, height, "#17A2B8", "#ffffff", text, link)
}

// TextButton renders a white button with colored text, like a hyperlink.
func TextButton(width, height int, text, link string) string {
	return Button(width, height, "#ffffff", "#007BFF", text, link)
}

// IconButton renders a square button with an icon character (e.g. emoji).
func IconButton(width, height int, icon, link string) string {
	return Button(width, height, "#E0E0E0", "#000000", icon, link)
}

// ButtonFactory provides a named-style constructor for buttons.
// Supported kinds: "primary", "danger", "success", "small", "wide", "text", "icon".
func ButtonFactory(kind string, width, height int, text, link string) string {
	switch kind {
	case "primary":
		return PrimaryButton(width, height, text, link)
	case "danger":
		return DangerButton(width, height, text, link)
	case "success":
		return SuccessButton(width, height, text, link)
	case "small":
		return SmallButton(width, height, text, link)
	case "wide":
		return WideButton(width, height, text, link)
	case "text":
		return TextButton(width, height, text, link)
	case "icon":
		return IconButton(width, height, text, link)
	default:
		return PrimaryButton(width, height, text, link)
	}
}
