package mermaid

import (
	"html"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/width"
)

// asciiWidths holds interWidths for ASCII, which most labels are, so
// measuring them is an array read rather than a map lookup.
var asciiWidths = func() (t [128][2]uint16) {
	for r, w := range interWidths {
		if r < 128 {
			t[r] = w
		}
	}
	return t
}()

// textWidth returns the width of s set in Inter at size px, semibold when
// bold. The widths come from the font gnoweb serves (widths_gen.go); a rune
// it lacks is set by the browser in a fallback font, estimated at a full em
// when it is wide and 0.6 em otherwise. Kerning is ignored: it moves a label
// by a pixel or two, which the padding absorbs.
func textWidth(s string, size float64, bold bool) float64 {
	var em float64
	for _, r := range s {
		em += runeEm(r, bold)
	}
	return em * size / 1000
}

// runeEm returns the width of r in thousandths of an em.
func runeEm(r rune, bold bool) float64 {
	weight := 0
	if bold {
		weight = 1
	}
	if r < 128 && asciiWidths[r][0] != 0 {
		return float64(asciiWidths[r][weight])
	}
	if w, ok := interWidths[r]; ok {
		return float64(w[weight])
	}
	switch {
	case unicode.In(r, unicode.Mn, unicode.Me, unicode.Cf):
		return 0 // combining marks and format characters take no room
	case wide(r):
		return 1000
	default:
		return 600
	}
}

// wide reports whether r is set a full em wide, as East Asian wide and
// fullwidth characters, emoji included, are.
func wide(r rune) bool {
	if r < 0x1100 { // no wide character comes before Hangul Jamo
		return false
	}
	k := width.LookupRune(r).Kind()
	return k == width.EastAsianWide || k == width.EastAsianFullwidth
}

func linesWidth(lines []string, size float64, bold bool) float64 {
	var w float64
	for _, l := range lines {
		w = max(w, textWidth(l, size, bold))
	}
	return w
}

// lineBreak is Mermaid's line break in a label.
var lineBreak = regexp.MustCompile(`(?i)<br\s*/?>`)

// label returns the lines of a label, or nil for an empty one: it splits the
// label at its <br> tags, as Mermaid does, and decodes the entity codes of
// each line (see decodeEntities), so an encoded tag stays text.
func label(s string) []string {
	if s = strings.TrimSpace(s); s == "" {
		return nil
	}
	if !strings.Contains(s, "<") {
		return []string{decodeEntities(s)}
	}
	lines := lineBreak.Split(s, -1)
	for i, l := range lines {
		lines[i] = decodeEntities(strings.TrimSpace(l))
	}
	return lines
}

// plain returns a title or a value as drawn: without the quotes around it,
// entity codes decoded.
func plain(s string) string { return decodeEntities(unquote(strings.TrimSpace(s))) }

func unquote(s string) string {
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') && s[len(s)-1] == s[0] {
		return s[1 : len(s)-1]
	}
	return s
}

func cutQuoted(s string) (inner, rest string, ok bool) {
	if !strings.HasPrefix(s, `"`) {
		return "", s, false
	}
	end := strings.IndexByte(s[1:], '"')
	if end < 0 {
		return "", s, false
	}
	return s[1 : end+1], strings.TrimSpace(s[end+2:]), true
}

// entityEnd returns the end of the entity code that starts at s[i], as in
// #59; or #quot;, or -1. It is Mermaid's syntax, #\w+;, whether or not the
// code is known.
func entityEnd(s string, i int) int {
	if i >= len(s) || s[i] != '#' {
		return -1
	}
	j := i + 1
	for j < len(s) && (isAlnum(s[j]) || s[j] == '_') {
		j++
	}
	if j == i+1 || j == len(s) || s[j] != ';' {
		return -1
	}
	return j + 1
}

func isAlnum(c byte) bool {
	return c >= '0' && c <= '9' || c|0x20 >= 'a' && c|0x20 <= 'z'
}

// decodeEntities decodes Mermaid's entity codes, which let a label hold the
// characters its syntax reserves: #59; for a code point and #quot; for an
// HTML entity name. A code for a control character, #10; or #NewLine;,
// decodes to a space, which is how the browser shows Mermaid's; an unknown
// or invalid code stays as it is. The decoded text is escaped like any other
// when written.
func decodeEntities(s string) string {
	if !strings.Contains(s, "#") {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		end := entityEnd(s, i)
		if end < 0 {
			b.WriteByte(s[i])
			continue
		}
		code := s[i+1 : end-1]
		decoded := html.UnescapeString("&" + code + ";")
		if n, err := strconv.ParseUint(code, 10, 32); err == nil {
			decoded = ""
			if n <= unicode.MaxRune {
				if r := rune(n); utf8.ValidRune(r) {
					decoded = string(r)
				}
			}
		} else if decoded == "&"+code+";" {
			decoded = "" // unknown
		}
		if decoded == "" {
			b.WriteByte(s[i])
			continue
		}
		for _, r := range decoded {
			if unicode.IsControl(r) {
				r = ' '
			}
			b.WriteRune(r)
		}
		i = end - 1
	}
	return b.String()
}

// wrap breaks each of lines at its spaces so it fits width when set at size. A
// word wider than width is broken too, after a / . - _ or , when it has one,
// and between any two characters otherwise, so an address, a path or a run
// of East Asian text never overflows.
func wrap(lines []string, size, width float64, bold bool) []string {
	var out []string
	space := textWidth(" ", size, bold)
	for _, l := range lines {
		var cur strings.Builder
		curW := 0.0
		// Words end at spaces only: a no-break space (#nbsp;) holds.
		for _, word := range strings.FieldsFunc(l, func(r rune) bool { return r == ' ' }) {
			ww := textWidth(word, size, bold)
			if ww > width {
				if cur.Len() > 0 {
					out = append(out, cur.String())
					cur.Reset()
				}
				pieces := breakWord(word, size, width, bold)
				out = append(out, pieces[:len(pieces)-1]...)
				word = pieces[len(pieces)-1]
				ww = textWidth(word, size, bold)
			}
			if cur.Len() > 0 && curW+space+ww > width {
				out = append(out, cur.String())
				cur.Reset()
			}
			if cur.Len() > 0 {
				cur.WriteByte(' ')
				curW += space
			} else {
				curW = 0
			}
			cur.WriteString(word)
			curW += ww
		}
		out = append(out, cur.String())
	}
	return out
}

// breakWord cuts word into pieces no wider than width, each but the last
// ending after a break character when one fits. It never breaks inside a
// character as drawn (see joins), so no piece starts with a lone accent,
// half an emoji or half a flag.
func breakWord(word string, size, width float64, bold bool) []string {
	var out []string
	start, w := 0, 0.0
	cut, cutW := -1, 0.0 // the last place to break, and the width before it
	prev := rune(-1)
	flags := 0 // the regional indicators just before r
	for i, r := range word {
		rw := runeEm(r, bold) * size / 1000
		if !joins(prev, r, flags) {
			if strings.ContainsRune("/.-_,", prev) || wide(prev) {
				cut, cutW = i, w
			}
			// A break at the cut may still leave too much: then break here too.
			for w+rw > width && i > start {
				if cut > start {
					out = append(out, word[start:cut])
					start, w = cut, w-cutW
				} else {
					out = append(out, word[start:i])
					start, w = i, 0
				}
				cut = -1
			}
		}
		w += rw
		prev = r
		if unicode.Is(unicode.Regional_Indicator, r) {
			flags++
		} else {
			flags = 0
		}
	}
	return append(out, word[start:])
}

// joins reports whether r continues the character drawn before it, so a
// word is not broken between them: r is a combining mark, a format
// character, an emoji skin tone, a Hangul vowel or final consonant, what
// follows a zero width joiner, or the second regional indicator of a flag,
// after an odd number of them.
func joins(prev, r rune, flags int) bool {
	return prev == '\u200d' ||
		unicode.In(r, unicode.Mn, unicode.Me, unicode.Mc, unicode.Cf) ||
		0x1f3fb <= r && r <= 0x1f3ff ||
		0x1160 <= r && r <= 0x11ff || 0xd7b0 <= r && r <= 0xd7ff ||
		unicode.Is(unicode.Regional_Indicator, r) && flags%2 == 1
}
