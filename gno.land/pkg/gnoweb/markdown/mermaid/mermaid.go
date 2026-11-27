// Package mermaid draws a subset of Mermaid diagrams (https://mermaid.js.org)
// as SVG, on the server, for gnoweb's realm markdown.
//
// The SVG is meant to be inlined in a page served under a strict
// Content-Security-Policy: it has no style attribute and no <style> element,
// only geometry and class names that gnoweb's stylesheet colors, so a diagram
// follows the light and dark themes and needs no script. Every element and
// attribute is written by svgWriter, which writes nothing else, and all text
// is escaped.
//
// Realm markdown is attacker input. A source longer than MaxSourceLen is not
// drawn, the layout of each diagram type takes time linear in its source, and
// a drawing larger than maxDimension or MaxOutputLen is refused, so the work
// and the output of a diagram stay bounded.
//
// Render returns an *Error for any source it does not draw, a statement
// outside the supported subset included: the caller then shows the source.
package mermaid

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// MaxSourceLen is the longest source Render draws, in bytes.
const MaxSourceLen = 16 << 10

// MaxOutputLen is the largest SVG Render returns, in bytes.
const MaxOutputLen = 512 << 10

// maxDimension is the widest and tallest drawing Render returns, in px.
const maxDimension = 20000

// Error says why a diagram was not drawn. Line is the 1-based source line it
// concerns, or 0.
type Error struct {
	Line int
	Msg  string
}

func (e *Error) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("mermaid: line %d: %s", e.Line, e.Msg)
	}
	return "mermaid: " + e.Msg
}

func errorf(line int, format string, args ...any) error {
	return &Error{Line: line, Msg: fmt.Sprintf(format, args...)}
}

// line is a source line: its number, its indentation in whitespace
// characters (a tab counts as one, as Mermaid counts it) and its text with
// the indentation and trailing spaces removed and its tabs made spaces, as
// Mermaid separates words with either.
type line struct {
	n      int
	indent int
	text   string
}

// meta holds what every diagram type may set: a title, which Render draws
// above the diagram, and a title and a description for assistive
// technology.
type meta struct {
	title, accTitle, accDescr string
}

// diagram is a parsed diagram.
type diagram interface {
	// draw writes the diagram, without its title, into w from 0, 0 and
	// returns its size, margins included.
	draw(w *svgWriter) (width, height float64)
}

// parseFunc parses the statements of a diagram. args is the rest of its
// header line, after the type keyword. m.title holds the title of the
// frontmatter, if any, which a title statement replaces.
type parseFunc func(args string, lines []line, m *meta) (diagram, error)

// kinds are the supported diagram types, by the keyword that opens them, with
// the name assistive technology reads when the diagram has no title.
var kinds = map[string]struct {
	name  string
	parse parseFunc
}{
	"sequenceDiagram": {"Sequence diagram", parseSequence},
	"pie":             {"Pie chart", parsePie},
	"xychart-beta":    {"Chart", parseXYChart},
	"xychart":         {"Chart", parseXYChart},
	"timeline":        {"Timeline", parseTimeline},
	"kanban":          {"Kanban board", parseKanban},
}

// Render draws a Mermaid source as an SVG element.
func Render(src string) (out []byte, err error) {
	defer func() {
		// The source is attacker input: a bug in a parser or a layout must
		// show the source, not take the page down.
		if r := recover(); r != nil {
			out, err = nil, errorf(0, "internal error: %v", r)
		}
	}()
	if len(src) > MaxSourceLen {
		return nil, errorf(0, "source longer than %d bytes", MaxSourceLen)
	}
	if !utf8.ValidString(src) {
		return nil, errorf(0, "source is not valid UTF-8")
	}
	var m meta
	lines, err := scan(src, &m)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, errorf(0, "empty diagram")
	}
	keyword, args, _ := strings.Cut(lines[0].text, " ")
	kind, ok := kinds[keyword]
	if !ok {
		return nil, errorf(lines[0].n, "unsupported diagram type %q", keyword)
	}
	body, err := accessibility(lines[1:], &m)
	if err != nil {
		return nil, err
	}
	d, err := kind.parse(strings.TrimSpace(args), body, &m)
	if err != nil {
		return nil, err
	}
	var w svgWriter
	bodyW, bodyH := d.draw(&w)
	title := plain(m.title)
	label := cmp.Or(plain(m.accTitle), title, kind.name)
	out, width, height := w.document(bodyW, bodyH, title, label, plain(m.accDescr))
	if width > maxDimension || height > maxDimension {
		return nil, errorf(0, "diagram larger than %d px", maxDimension)
	}
	if len(out) > MaxOutputLen {
		return nil, errorf(0, "diagram larger than %d bytes", MaxOutputLen)
	}
	return out, nil
}

// scan splits src into lines, dropping blank lines, %% comments and the
// %%{init: …}%% directives, which may span lines (a diagram is drawn with
// gnoweb's theme, whatever configuration it asks for), and a leading YAML
// frontmatter, whose title it keeps in m.
func scan(src string, m *meta) ([]line, error) {
	var out []line
	raw := strings.Split(src, "\n")
	i := 0
	for i < len(raw) && strings.TrimSpace(raw[i]) == "" {
		i++
	}
	if i < len(raw) && strings.TrimSpace(raw[i]) == "---" {
		start := i
		for i++; i < len(raw) && strings.TrimSpace(raw[i]) != "---"; i++ {
			// Only a top-level key: config may nest a title of its own.
			if v, ok := statement(strings.TrimRight(raw[i], " \t\r"), "title"); ok {
				m.title = v
			}
		}
		if i == len(raw) {
			return nil, errorf(start+1, "frontmatter is not closed")
		}
		i++
	}
	for ; i < len(raw); i++ {
		text := strings.TrimRight(raw[i], " \t\r")
		indent := len(text) - len(strings.TrimLeft(text, " \t"))
		text = strings.ReplaceAll(text[indent:], "\t", " ")
		if strings.HasPrefix(text, "%%{") {
			for !strings.HasSuffix(strings.TrimSpace(raw[i]), "}%%") {
				if i++; i == len(raw) {
					return nil, errorf(0, "directive is not closed")
				}
			}
			continue
		}
		if text == "" || strings.HasPrefix(text, "%%") {
			continue
		}
		out = append(out, line{n: i + 1, indent: indent, text: text})
	}
	return out, nil
}

// accessibility takes the accTitle and accDescr statements, which every
// diagram type accepts, out of lines.
func accessibility(lines []line, m *meta) ([]line, error) {
	out := lines[:0:0]
	for i := 0; i < len(lines); i++ {
		l := lines[i]
		if v, ok := statement(l.text, "accTitle"); ok {
			m.accTitle = v
			continue
		}
		if v, ok := statement(l.text, "accDescr"); ok {
			m.accDescr = v
			continue
		}
		rest, ok := cutPrefixFold(l.text, "accDescr")
		rest = strings.TrimSpace(rest)
		if inner, closed := strings.CutSuffix(rest, "}"); ok && closed && strings.HasPrefix(inner, "{") {
			m.accDescr = strings.TrimSpace(inner[1:]) // accDescr { on one line }
			continue
		}
		if ok && rest == "{" {
			var desc []string
			for i++; i < len(lines) && lines[i].text != "}"; i++ {
				desc = append(desc, lines[i].text)
			}
			if i == len(lines) {
				return nil, errorf(l.n, "accDescr block is not closed")
			}
			m.accDescr = strings.Join(desc, " ")
			continue
		}
		out = append(out, l)
	}
	return out, nil
}

// lowerASCII lowercases the ASCII letters of s and only them: Mermaid's
// keywords ignore ASCII case, so "rİght" is not "right".
func lowerASCII(s string) string {
	return strings.Map(func(r rune) rune {
		if 'A' <= r && r <= 'Z' {
			return r + 'a' - 'A'
		}
		return r
	}, s)
}

// equalFold reports whether s and t are equal ignoring ASCII case.
func equalFold(s, t string) bool { return len(s) == len(t) && lowerASCII(s) == lowerASCII(t) }

// cutPrefixFold is strings.CutPrefix ignoring ASCII case, as Mermaid's
// keywords do.
func cutPrefixFold(s, prefix string) (string, bool) {
	if len(s) < len(prefix) || !equalFold(s[:len(prefix)], prefix) {
		return s, false
	}
	return s[len(prefix):], true
}

// statement reports whether text is `keyword: value` and returns the value.
func statement(text, keyword string) (string, bool) {
	rest, ok := cutPrefixFold(text, keyword)
	if !ok {
		return "", false
	}
	rest = strings.TrimSpace(rest)
	if !strings.HasPrefix(rest, ":") {
		return "", false
	}
	return strings.TrimSpace(rest[1:]), true
}

// title reports whether text is a title statement, `title Text` or
// `title: Text`, and returns the title.
func title(text string) (string, bool) {
	if v, ok := statement(text, "title"); ok {
		return v, true
	}
	if rest, ok := cutPrefixFold(text, "title "); ok {
		return strings.TrimSpace(rest), true
	}
	return "", false
}

func firstWord(s string) string {
	w, _, _ := strings.Cut(s, " ")
	return w
}

// parseNumber parses a decimal number of at most 1e15 in magnitude, in the
// syntax Mermaid reads: digits, a point, a sign and an exponent, which rules
// out the hexadecimal, the underscores and the infinities strconv accepts.
func parseNumber(s string) (float64, error) {
	if s == "" || strings.Trim(s, "0123456789.eE+-") != "" {
		return 0, strconv.ErrSyntax
	}
	v, err := strconv.ParseFloat(s, 64)
	if err == nil && (v > 1e15 || v < -1e15) {
		err = strconv.ErrRange
	}
	return v, err
}
