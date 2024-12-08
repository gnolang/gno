package logger

import (
	"fmt"
	"hash/fnv"
	"math"

	"github.com/charmbracelet/lipgloss"
)

func colorFromString(s string, saturation, lightness float64) lipgloss.Color {
	hue := float64(hash32a(s) % 360)

	r, g, b := hslToRGB(hue, saturation, lightness)
	hex := rgbToHex(r, g, b)
	return lipgloss.Color(hex)
}

func hash32a(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// from: https://www.rapidtables.com/convert/color/hsl-to-rgb.html
// hslToRGB converts an HSL triple to an RGB triple.
func hslToRGB(h, s, l float64) (r, g, b uint8) {
	if h < 0 || h >= 360 || s < 0 || s > 1 || l < 0 || l > 1 {
		return 0, 0, 0
	}

	C := (1 - math.Abs((2*l)-1)) * s
	X := C * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - (C / 2)

	var rNot, gNot, bNot float64
	switch {
	case 0 <= h && h < 60:
		rNot, gNot, bNot = C, X, 0
	case 60 <= h && h < 120:
		rNot, gNot, bNot = X, C, 0
	case 120 <= h && h < 180:
		rNot, gNot, bNot = 0, C, X
	case 180 <= h && h < 240:
		rNot, gNot, bNot = 0, X, C
	case 240 <= h && h < 300:
		rNot, gNot, bNot = X, 0, C
	case 300 <= h && h < 360:
		rNot, gNot, bNot = C, 0, X
	}

	r = uint8(math.Round((rNot + m) * 255))
	g = uint8(math.Round((gNot + m) * 255))
	b = uint8(math.Round((bNot + m) * 255))
	return r, g, b
}

func rgbToHex(r, g, b uint8) string {
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}
