package gnolang

import (
	"fmt"
)

type runestate int

const (
	runestateCode           runestate = 0
	runestateRune           runestate = 1
	runestateStringQuote    runestate = 2
	runestateStringBacktick runestate = 3
)

type scanner struct {
	str string
	rnz []rune
	idx int
	runestate
	curly  int
	round  int
	square int
}

// returns a new scanner.
func newScanner(str string) *scanner {
	rnz := make([]rune, 0, len(str))
	for _, r := range str {
		rnz = append(rnz, r)
	}
	return &scanner{
		str:       str,
		runestate: runestateCode,
		rnz:       rnz,
	}
}

// Peeks the next n runes and returns a string.  returns a shorter string if
// there are less than n runes left.
func (ss *scanner) peek(n int) string {
	if ss.idx+n > len(ss.rnz) {
		return string(ss.rnz[ss.idx:len(ss.rnz)])
	}
	return string(ss.rnz[ss.idx : ss.idx+n])
}

// Advance a single rune, e.g. by incrementing ss.curly if ss.rnz[ss.idx] is
// '{' before advancing.  If ss.runestate is runestateRune or runestateQuote,
// advances escape sequences to completion so ss.idx may increment more than
// one.  Returns true if done.
func (ss *scanner) advance() bool {
	rn := ss.rnz[ss.idx] // just panic if out of scope, caller error.
	switch ss.runestate {
	case runestateCode:
		switch rn {
		case '}':
			ss.curly--
			if ss.curly < 0 {
				panic("mismatched curly: " + ss.str)
			}
		case ')':
			ss.round--
			if ss.round < 0 {
				panic("mismatched round: " + ss.str)
			}
		case ']':
			ss.square--
			if ss.square < 0 {
				panic("mismatched square: " + ss.str)
			}
		case '{':
			ss.curly++
		case '(':
			ss.round++
		case '[':
			ss.square++
		case '\'':
			ss.runestate = runestateRune
		case '"':
			ss.runestate = runestateStringQuote
		case '`':
			ss.runestate = runestateStringBacktick
		}
	case runestateRune:
		switch rn {
		case '\\':
			return ss.advanceEscapeSequence()
		case '\'':
			ss.runestate = runestateCode
		}
	case runestateStringQuote:
		switch rn {
		case '\\':
			return ss.advanceEscapeSequence()
		case '"':
			ss.runestate = runestateCode
		}
	case runestateStringBacktick:
		switch rn {
		case '`':
			ss.runestate = runestateCode
		}
	}
	ss.idx++
	return ss.done()
}

// returns true if no runes left to advance.
func (ss *scanner) done() bool {
	return ss.idx == len(ss.rnz)
}

// returns true if outside the scope of any
// parentheses, brackets, strings, or rune literals.
func (ss *scanner) out() bool {
	return ss.runestate == runestateCode &&
		ss.curly == int(0) &&
		ss.round == int(0) &&
		ss.square == int(0)
}

func isOctal(r rune) bool {
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7':
		return true
	default:
		return false
	}
}

func isHex(r rune) bool {
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'a', 'b', 'c', 'd', 'e', 'f',
		'A', 'B', 'C', 'D', 'E', 'F':
		return true
	default:
		return false
	}
}

// Advances runes, while checking that each passes `check`.  if error, panics
// with info including `back` runes back.
func (ss *scanner) eatRunes(back int, eat int, check func(rune) bool) {
	for i := 0; i < eat; i++ {
		if ss.idx+i == len(ss.rnz) {
			panic(fmt.Sprintf("eof while parsing: %s",
				string(ss.rnz[ss.idx-back:])))
		}
		if !check(ss.rnz[ss.idx+i]) {
			panic(fmt.Sprintf("invalid character while parsing: %s",
				string(ss.rnz[ss.idx-back:ss.idx+i+1])))
		}
		ss.idx++
	}
}

// increments ss.idx until escape sequence is complete.  returns true if done.
func (ss *scanner) advanceEscapeSequence() bool {
	rn1 := ss.rnz[ss.idx]
	if rn1 != '\\' {
		panic("should not happen")
	}
	if ss.idx == len(ss.rnz)-1 {
		panic("eof while parsing escape sequence")
	}
	rn2 := ss.rnz[ss.idx+1]
	switch rn2 {
	case 'x':
		ss.idx += 2
		ss.eatRunes(2, 2, isHex)
		return ss.done()
	case 'u':
		ss.idx += 2
		ss.eatRunes(2, 4, isHex)
		return ss.done()
	case 'U':
		ss.idx += 2
		ss.eatRunes(2, 8, isHex)
		return ss.done()
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', '\'', '"':
		ss.idx += 2
		return ss.done()
	default:
		ss.idx += 1
		if isOctal(rn2) {
			ss.eatRunes(1, 3, isOctal)
		} else {
			panic("invalid escape sequence")
		}
		return ss.done()
	}
}
