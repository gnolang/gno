package rawterm

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
	KeyCtrlS KeyPress = '\x13' // Ctrl+S
	KeyCtrlT KeyPress = '\x14' // Ctrl+T

	KeyA KeyPress = 'A'
	KeyE KeyPress = 'E'
	KeyH KeyPress = 'H'
	KeyI KeyPress = 'I'
	KeyN KeyPress = 'N'
	KeyP KeyPress = 'P'
	KeyR KeyPress = 'R'

	// Special keys
	KeyUp    KeyPress = 0x80 // Arbitrary value outside ASCII range
	KeyDown  KeyPress = 0x81
	KeyLeft  KeyPress = 0x82
	KeyRight KeyPress = 0x83
)

func (k KeyPress) Upper() KeyPress {
	return KeyPress(unicode.ToUpper(rune(k)))
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
	case KeyCtrlS:
		return "Ctrl+S"
	case KeyCtrlT:
		return "Ctrl+T"
	case KeyUp:
		return "Up Arrow"
	case KeyDown:
		return "Down Arrow"
	case KeyLeft:
		return "Left Arrow"
	case KeyRight:
		return "Right Arrow"
	default:
		// For printable ASCII characters
		if k > 0x20 && k < 0x7e {
			return fmt.Sprintf("%c", k)
		}

		return fmt.Sprintf("Unknown (0x%02x)", byte(k))
	}
}
