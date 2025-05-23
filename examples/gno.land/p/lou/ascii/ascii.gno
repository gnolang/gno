package ascii

import (
	"strconv"
	"strings"

	"gno.land/p/moul/md"
)

type WAlignment string
type HAlignment string

const (
	// Width Alignment
	AlignWCenter WAlignment = "center"
	AlignLeft    WAlignment = "left"
	AlignRight   WAlignment = "right"

	// Height Alignment
	AlignHCenter HAlignment = "center"
	AlignTop     HAlignment = "top"
	AlignBottom  HAlignment = "bottom"
)

// padLine aligns text within a given width.
//
// Supports AlignLeft, AlignRight, and AlignWCenter alignment.
func padLine(line string, width int, align WAlignment, space string) string {
	padding := width - len(line)
	if width < len(line) {
		padding = 0
	}

	switch align {
	case AlignRight:
		return Repeat(space, padding) + line
	case AlignWCenter:
		left := padding / 2
		right := padding - left
		return Repeat(space, left) + line + Repeat(space, right)
	default: // AlignLeft
		return line + Repeat(space, padding)
	}
}

// padHeight pads lines vertically according to alignment.
//
// Supports AlignTop, AlignBottom, and AlignHCenter alignment.
func padHeight(lines []string, height int, align HAlignment) []string {
	padded := []string{}
	if height <= 0 {
		return lines
	}
	extra := height - len(lines)

	topPad := 0
	switch align {
	case AlignBottom:
		topPad = extra
	case AlignHCenter:
		topPad = extra / 2
	}

	for i := 0; i < topPad; i++ {
		padded = append(padded, "")
	}
	padded = append(padded, lines...)
	for len(padded) < height {
		padded = append(padded, "")
	}
	return padded
}

// Repeat returns repetition of a string n times.
func Repeat(char string, n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(char, n)
}

// Box draws a single-line text in a simple box.
// If the string contains newlines, it falls back to FlexFrame function.
//
// Example:
//
//	Box("Hello World\n!")
//
// Gives:
//
//	+-------------+
//	| Hello World |
//	| !           |
//	+-------------+
func Box(text string) string {
	return FlexFrame(strings.Split(text, "\n"), AlignLeft)
}

// FlexFrame draws a frame with automatic width and alignment.
//
// Example:
//
//	FlexFrame([]string{"hello", "worldd", "!!"}, "right")
//
// Gives:
//
//	+-------+
//	|  hello |
//	| worldd |
//	|     !! |
//	+-------+
func FlexFrame(lines []string, align WAlignment) string {
	maxWidth := 0

	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	top := "+" + Repeat("-", maxWidth+2) + "+\n"
	bottom := "+" + Repeat("-", maxWidth+2) + "+"

	body := ""
	for i := 0; i < len(lines); i++ {
		body += "| " + padLine(lines[i], maxWidth, align, " ") + " |\n"
	}

	return md.CodeBlock(top+body+bottom) + "\n"
}

// Frame draws a frame with specific width, height and alignment options.
//
// Example:
//
//	Frame([]string{"hello", "world", "!!"}, "center", 10, 5, "center")
//
// Gives:
//
//	+------------+
//	|            |
//	|   hello    |
//	|   world    |
//	|     !!     |
//	|            |
//	+------------+
func Frame(
	lines []string,
	wAlign WAlignment,
	width, height int,
	hAlign HAlignment,
) string {
	maxWidth := width
	for i := 0; i < len(lines); i++ {
		if len(lines[i]) > maxWidth {
			maxWidth = len(lines[i])
		}
	}

	if len(lines) > height {
		height = len(lines)
	}
	lines = padHeight(lines, height, hAlign)

	top := "+" + Repeat("-", maxWidth+2) + "+\n"
	bottom := "+" + Repeat("-", maxWidth+2) + "+"
	body := ""

	for _, line := range lines {
		body += "| " + padLine(line, maxWidth, wAlign, " ") + " |\n"
	}

	return md.CodeBlock(top+body+bottom) + "\n"
}

// FixedFrame draws a frame with a fixed width and height, truncating or wrapping content as needed.
// Width and height include the content area, not the frame borders.
//
// Example:
//
//	Frame([]string{"hello world!!"}, ascii.AlignWCenter, 10, 4, ascii.AlignHCenter)
//
// Gives:
//
//	+------------+
//	|            |
//	|   hello    |
//	|  world!!   |
//	|            |
//	+------------+
func FixedFrame(
	lines []string,
	wAlign WAlignment,
	width, height int,
	hAlign HAlignment,
) string {
	var wrapped []string
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	for _, line := range lines {
		words := strings.Fields(line)
		current := ""
		for _, word := range words {
			if len(current)+len(word)+1 > width {
				wrapped = append(wrapped, current)
				current = word
			} else {
				if current == "" {
					current = word
				} else {
					current += " " + word
				}
			}
		}
		if current != "" {
			wrapped = append(wrapped, current)
		}
	}
	wrapped = padHeight(wrapped, height, hAlign)

	top := "+" + Repeat("-", width+2) + "+\n"
	bottom := "+" + Repeat("-", width+2) + "+"

	body := ""
	for i, line := range wrapped {
		if i == height {
			break
		}
		body += "| " + padLine(line, width, wAlign, " ") + " |\n"
	}

	return md.CodeBlock(top+body+bottom) + "\n"
}

// ProgressBar renders a visual progress bar, the size represents the number of chars in length for the bar.
//
// Example:
//
//	ProgressBar(2, 6, 10, true)
//
// Gives: [###-------] 33%
func ProgressBar(current int, total int, charSize int, displayPercent bool) string {
	if total == 0 {
		return PercentageBar(0, charSize, displayPercent)
	}
	percent := (current * 100) / total

	return PercentageBar(percent, charSize, displayPercent)
}

// PercentageBar renders a visual progress bar, the size represents the number of chars in length for the bar.
// This differs from ProgressBar in that it does not require a total value, takes a percentage directly.
//
// Example:
//
//	PercentageBar(50, 6, true)
//
// Gives: [###---] 50%
func PercentageBar(percent int, charSize int, displayPercent bool) string {
	fillLength := (percent * charSize) / 100
	emptyLength := charSize - fillLength

	filled := Repeat("#", fillLength)
	empty := Repeat("-", emptyLength)

	out := "[" + filled + empty + "]"
	if !displayPercent {
		return out
	}
	return out + " " + strconv.Itoa(percent) + "%"
}

// Grid renders a 2D grid of characters.
//
// Example:
//
//	Grid(3, 3, "x")
//
// Gives:
//
//	xxx
//	xxx
//	xxx
func Grid(rows int, cols int, fill string) string {
	out := ""
	if rows <= 0 || cols <= 0 {
		return out
	}

	for r := 0; r < rows; r++ {
		row := ""
		for c := 0; c < cols; c++ {
			row += fill
		}
		out += row + "\n"
	}
	return out
}
