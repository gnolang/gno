package dev

import (
	"fmt"
	"unicode"
)

type KeyPress byte

// key representation
const (
	KeyNone  KeyPress = 0      // None
	KeyCtrlC KeyPress = '\x03' // Ctrl+C
	KeyCtrlD KeyPress = '\x04' // Ctrl+D
	KeyCtrlE KeyPress = '\x05' // Ctrl+E
	KeyCtrlL KeyPress = '\x0c' // Ctrl+L
	KeyCtrlO KeyPress = '\x0f' // Ctrl+O
	KeyCtrlR KeyPress = '\x12' // Ctrl+R
	KeyCtrlT KeyPress = '\x14' // Ctrl+T
)

const (
	// ANSI escape codes
	ClearCurrentLine = "\033[2K"
	MoveCursorUp     = "\033[1A"
	MoveCursorDown   = "\033[1B"
)

func (k KeyPress) Lower() KeyPress {
	return KeyPress(unicode.ToLower(rune(k)))
}

func (k KeyPress) String() string {
	switch k {
	case KeyNone:
		return "Null"
	case KeyCtrlC:
		return "Ctrl+C"
	case KeyCtrlD:
		return "Ctrl+D"
	case KeyCtrlE:
		return "Ctrl+E"
	case KeyCtrlL:
		return "Ctrl+L"
	case KeyCtrlO:
		return "Ctrl+O"
	case KeyCtrlR:
		return "Ctrl+R"
	case KeyCtrlT:
		return "Ctrl+T"
		// For printable ASCII characters
	default:
		if k > 0x20 && k < 0x7e {
			return fmt.Sprintf("%c", k)
		}

		return fmt.Sprintf("Unknown (0x%02x)", byte(k))
	}
}
