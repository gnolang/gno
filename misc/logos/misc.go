package logos

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	runewidth "github.com/mattn/go-runewidth"
)

// splits a string into lines by newline.
func splitLines(s string) (ss []string) {
	return strings.Split(s, "\n")
}

// splits a string according to unicode spaces.
func splitSpaces(s string) (ss []string) {
	buf := []rune{}
	for _, r := range s {
		if unicode.IsSpace(r) {
			// continue
			if len(buf) > 0 {
				ss = append(ss, string(buf))
				buf = nil
			}
		} else {
			buf = append(buf, r)
		}
	}
	if len(buf) > 0 {
		ss = append(ss, string(buf))
		// buf = nil
	}
	return ss
}

func toRunes(s string) []rune {
	runes := make([]rune, 0, len(s))
	for _, r := range s {
		runes = append(runes, r)
	}
	return runes
}

// gets the terminal display width of a string.
// must be compatible with nextCharacter().
// NOTE: must be kept in sync with nextCharacter(); see tests.
func widthOf(s string) (l int) {
	zwj := false // zero width joiner '\u200d'.
	for _, r := range s {
		if r == '\u200d' {
			zwj = true
			continue
		}
		if zwj {
			zwj = false
			continue
		}
		switch runewidth.RuneWidth(r) {
		case 0:
			if isCombining(r) {
				// combining characters have no length.
			} else {
				l++ // show a blank instead, weird.
			}
		case 1:
			l++
		case 2:
			l += 2
		default:
			panic("should not happen")
		}
	}
	return l
}

// given runes of a valid utf8 string,
// return a string that represents
// the next single character (with any modifiers).
// w: width of character. n: number of runes read
func nextCharacter(rz []rune) (s string, w int, n int) {
	for n = 0; n < len(rz); n++ {
		r := rz[n]
		if r == '\u200d' {
			// special case: zero width joins.
			s = s + string(r)
			if n+1 < len(rz) {
				s = s + string(rz[n+1])
				n++
				continue
			} else {
				// just continue, return invalid string s.
				n++
				return
			}
		} else if 0 < len(s) {
			return
		} else {
			// append r to s and inc w.
			rw := runewidth.RuneWidth(r)
			s = s + string(r)
			if rw == 0 {
				if isCombining(r) {
					// no width
				} else {
					w += 1
				}
			} else {
				w += rw
			}
		}
	}
	return
}

//----------------------------------------

func AbsCoord(elem Elem) (crd Coord) {
	for elem != nil {
		crd = crd.Add(elem.GetCoord())
		elem = elem.GetParent()
	}
	return
}

var randColors []Color = []Color{
	tcell.ColorAliceBlue,
	tcell.ColorAntiqueWhite,
	tcell.ColorAquaMarine,
	tcell.ColorAzure,
	tcell.ColorBeige,
	tcell.ColorBisque,
	tcell.ColorBlanchedAlmond,
	tcell.ColorBlueViolet,
	tcell.ColorBrown,
	tcell.ColorBurlyWood,
}

var rctr = 0

func RandColor() Color {
	rctr++
	return randColors[rctr%len(randColors)]
}

func IsInBounds(x, y int, origin Coord, size Size) bool {
	if x < origin.X || y < origin.Y {
		return false
	}
	if origin.X+size.Width <= x ||
		origin.Y+size.Height <= y {
		return false
	}
	return true
}
