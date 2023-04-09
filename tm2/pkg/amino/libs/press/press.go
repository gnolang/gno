package press

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino/libs/detrand"
)

type line struct {
	indentStr string // the fully expanded indentation string
	value     string // the original contents of the line
}

func newLine(indentStr string, value string) line {
	return line{indentStr, value}
}

func (l line) String() string {
	return l.indentStr + l.value
}

// Press is a tool for printing code.
// Press is not concurrency safe.
type Press struct {
	rnd          *rand.Rand // for generating a random variable names.
	indentPrefix string     // current indent prefix.
	indentDelim  string     // a tab or spaces, whatever.
	newlineStr   string     // should probably just remain "\n".
	lines        []line     // accumulated lines from printing.
}

func NewPress() *Press {
	return &Press{
		rnd:          rand.New(rand.NewSource(0)), //nolint:gosec
		indentPrefix: "",
		indentDelim:  "\t",
		newlineStr:   "\n",
		lines:        nil,
	}
}

func (p *Press) SetIndentDelim(s string) *Press {
	p.indentDelim = s
	return p
}

func (p *Press) SetNewlineStr(s string) *Press {
	p.newlineStr = s
	return p
}

// Main function for printing something on the press.
func (p *Press) P(s string, args ...interface{}) *Press {
	var l *line
	if len(p.lines) == 0 {
		// Make a new line.
		p.lines = []line{newLine(p.indentPrefix, "")}
	}
	// Get ref to last line.
	l = &(p.lines[len(p.lines)-1])
	l.value += fmt.Sprintf(s, args...)
	return p
}

// Appends a new line.
// It is also possible to print newline characters directly,
// but Press doesn't treat them as newlines for the sake of indentation.
func (p *Press) Ln() *Press {
	p.lines = append(p.lines, newLine(p.indentPrefix, ""))
	return p
}

// Convenience for P() followed by Nl().
func (p *Press) Pl(s string, args ...interface{}) *Press {
	return p.P(s, args...).Ln()
}

// auto-indents p2, appends concents to p.
// Panics if the last call wasn't Pl() or Ln().
// Regardless of whether Pl or Ln is called on p2,
// the indented lines terminate with newlineDelim before
// the next unindented line.
func (p *Press) I(block func(p2 *Press)) *Press {
	if len(p.lines) > 0 {
		lastLine := p.lines[len(p.lines)-1]
		if lastLine.value != "" {
			panic("cannot indent after nonempty line")
		}
		if lastLine.indentStr != p.indentPrefix {
			panic("unexpected indent string in last line")
		}
		// remove last empty line
		p.lines = p.lines[:len(p.lines)-1]
	}
	p2 := p.SubPress()
	p2.indentPrefix = p.indentPrefix + p.indentDelim
	block(p2)
	ilines := p2.Lines()
	// remove last empty line from p2
	ilines = withoutFinalNewline(ilines)
	p.lines = append(p.lines, ilines...)
	// (re)introduce last line with original indent
	p.lines = append(p.lines, newLine(p.indentPrefix, ""))
	return p
}

// Prints the final representation of the contents.
func (p *Press) Print() string {
	lines := []string{}
	for _, line := range p.lines {
		lines = append(lines, line.String())
	}
	return strings.Join(lines, p.newlineStr)
}

// Returns the lines.
// This may be useful for adding additional indentation to each line for code blocks.
func (p *Press) Lines() (lines []line) {
	return p.lines
}

// Convenience
func (p *Press) RandID(prefix string) string {
	return prefix + "_" + p.RandStr(8)
}

// Convenience
func (p *Press) RandStr(length int) string {
	const strChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" // 62 characters
	chars := []byte{}
MAIN_LOOP:
	for {
		val := p.rnd.Int63()
		for i := 0; i < 10; i++ {
			v := int(val & 0x3f) // rightmost 6 bits
			if v >= 62 {         // only 62 characters in strChars
				val >>= 6
				continue
			} else {
				chars = append(chars, strChars[v])
				if len(chars) == length {
					break MAIN_LOOP
				}
				val >>= 6
			}
		}
	}
	return string(chars)
}

// SubPress creates a blank Press suitable for inlining code.
// It starts with no indentation, zero lines,
// a derived rand from the original, but the same indent and nl strings..
func (p *Press) SubPress() *Press {
	p2 := NewPress()
	p2.rnd = detrand.DeriveRand(p.rnd)
	p2.indentPrefix = ""
	p2.indentDelim = p.indentDelim
	p2.newlineStr = p.newlineStr
	p2.lines = nil
	return p2
}

// ref: the reference to the value being encoded.
type EncoderPressFunc func(p *Press, ref string) (code string)

// ----------------------------------------

// If the final line is a line with no value, remove it
func withoutFinalNewline(lines []line) []line {
	if len(lines) > 0 && lines[len(lines)-1].value == "" {
		return lines[:len(lines)-1]
	} else {
		return lines
	}
}
