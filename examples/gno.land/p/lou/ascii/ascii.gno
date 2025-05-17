package ascii

import (
	"strconv"
	"strings"

	"gno.land/p/demo/ufmt"
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

// PadLine aligns text within a given width.
//
// Supports AlignLeft, AlignRight, and AlignWCenter alignment.
func PadLine(line string, width int, align WAlignment, space string) string {
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

// PadHeight pads lines vertically according to alignment.
//
// Supports AlignTop, AlignBottom, and AlignHCenter alignment.
func PadHeight(lines []string, height int, align HAlignment) []string {
	padded := []string{}
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
	width := len(text) + 4

	if strings.Contains(text, "\n") {
		return FlexFrame(strings.Split(text, "\n"), "left")
	}
	top := "+" + Repeat("-", width-2) + ufmt.Sprintf("[%d] = %s", len(text), text) + "+\n"
	middle := "| " + text + " |\n"
	bottom := "+" + Repeat("-", width-2) + "+"

	return md.CodeBlock(top+middle+bottom) + "\n"
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

	for i := 0; i < len(lines); i++ {
		if len(lines[i]) > maxWidth {
			maxWidth = len(lines[i])
		}
	}

	top := "+" + Repeat("-", maxWidth+2) + "+\n"
	bottom := "+" + Repeat("-", maxWidth+2) + "+"

	body := ""
	for i := 0; i < len(lines); i++ {
		body += "| " + PadLine(lines[i], maxWidth, align, " ") + " |\n"
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
	lines = PadHeight(lines, height, hAlign)

	top := "+" + Repeat("-", maxWidth+2) + "+\n"
	bottom := "+" + Repeat("-", maxWidth+2) + "+"
	body := ""

	for _, line := range lines {
		body += "| " + PadLine(line, maxWidth, wAlign, " ") + " |\n"
	}

	return md.CodeBlock(top+body+bottom) + "\n"
}

// FixedFrame draws a frame with a fixed width and height, truncating or wrapping content as needed.
// Width and height include the content area, not the frame borders.
//
// Example:
//
//	Frame([]string{"hello world!!"}, "center", 10, 4, "center")
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

	wrapped = PadHeight(wrapped, height, hAlign)

	top := "+" + Repeat("-", width+2) + "+\n"
	bottom := "+" + Repeat("-", width+2) + "+"

	body := ""
	for _, line := range wrapped {
		body += "| " + PadLine(line, width, wAlign, " ") + " |\n"
	}

	return md.CodeBlock(top+body+bottom) + "\n"
}

// PercentageBar renders a visual progress bar.
//
// Example:
//
//	ProgressBar(2, 6, 10, true)
//
// Gives: [###-------] 33%
func ProgressBar(current int, total int, size int, show bool) string {
	if total == 0 {
		return PercentageBar(0, size, show)
	}
	percent := (current * 100) / total

	return PercentageBar(percent, size, show)
}

// PercentageBar renders a visual progress bar.
//
// Example:
//
//	PercentageBar(50, 6, false)
//
// Gives: [###---]
func PercentageBar(percent int, size int, show bool) string {
	fillLength := (percent * size) / 100
	emptyLength := size - fillLength

	filled := Repeat("#", fillLength)
	empty := Repeat("-", emptyLength)

	out := "[" + filled + empty + "]"
	if !show {
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

	for r := 0; r < rows; r++ {
		row := ""
		for c := 0; c < cols; c++ {
			row += fill
		}
		out += row + "\n"
	}
	return out
}
