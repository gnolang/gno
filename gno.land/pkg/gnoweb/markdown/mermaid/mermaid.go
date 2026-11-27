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
	"math"
	"regexp"
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

type diagram interface {
	// draw writes the diagram, without its title, into w from 0, 0 and
	// returns its size, margins included.
	draw(w *svgWriter) (width, height float64)
}

// parseFunc parses the statements of a diagram. args is the rest of its
// header line, after the type keyword. m.title holds the title as drawn:
// the frontmatter's, if any, which a title statement replaces.
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

// Render draws a Mermaid source as an SVG element and returns it with its
// width, in px.
func Render(src string) (svg []byte, width float64, err error) {
	defer func() {
		// The source is attacker input: a bug in a parser or a layout must
		// show the source, not take the page down.
		if r := recover(); r != nil {
			svg, width, err = nil, 0, errorf(0, "internal error: %v", r)
		}
	}()
	if len(src) > MaxSourceLen {
		return nil, 0, errorf(0, "source longer than %d bytes", MaxSourceLen)
	}
	if !utf8.ValidString(src) {
		return nil, 0, errorf(0, "source is not valid UTF-8")
	}
	var m meta
	lines, err := scan(src, &m)
	if err != nil {
		return nil, 0, err
	}
	if len(lines) == 0 {
		return nil, 0, errorf(0, "empty diagram")
	}
	keyword, args, _ := strings.Cut(lines[0].text, " ")
	kind, ok := kinds[keyword]
	if !ok {
		return nil, 0, errorf(lines[0].n, "unsupported diagram type %q", keyword)
	}
	body, err := accessibility(lines[1:], &m)
	if err != nil {
		return nil, 0, err
	}
	d, err := kind.parse(strings.TrimSpace(args), body, &m)
	if err != nil {
		return nil, 0, err
	}
	var w svgWriter
	bodyW, bodyH := d.draw(&w)
	title := m.title
	// Checked before document copies the body, which its buffer stopped
	// writing past MaxOutputLen.
	width, height := max(bodyW, titleWidth(title)), bodyH+titleHeight(title)
	if width > maxDimension || height > maxDimension {
		return nil, 0, errorf(0, "diagram larger than %d px", maxDimension)
	}
	if w.b.Len() > MaxOutputLen {
		return nil, 0, errorf(0, "diagram larger than %d bytes", MaxOutputLen)
	}
	label := cmp.Or(plain(m.accTitle), title, kind.name)
	svg = w.document(width, height, bodyW, title, label, plain(m.accDescr))
	if len(svg) > MaxOutputLen {
		return nil, 0, errorf(0, "diagram larger than %d bytes", MaxOutputLen)
	}
	return svg, width, nil
}

// scan splits src into lines, dropping blank lines, %% comments and the
// %%{init: …}%% directives, which may span lines (a diagram is drawn with
// gnoweb's theme, whatever configuration it asks for), and a leading YAML
// frontmatter, whose title it keeps in m as YAML reads it. Mermaid reads the
// frontmatter before the entity codes, so the title's stay as written.
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
			key := frontTitle.FindStringSubmatch(strings.TrimRight(raw[i], " \t\r"))
			if key == nil {
				continue
			}
			t, ok := yamlScalar(key[1])
			if !ok || i+1 < len(raw) && continues(raw[i+1]) {
				return nil, errorf(i+1, "unsupported frontmatter title")
			}
			m.title = t
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

// frontTitle matches the title key of a frontmatter line and its value.
// YAML keys are case-sensitive, and a key needs a space after its colon.
var frontTitle = regexp.MustCompile(`^title[ \t]*:(?:[ \t](.*))?$`)

// yamlTyped matches the plain values YAML's core schema, as js-yaml reads
// it, takes for a bool or a number, binary and underscores in the digits
// included; js-yaml reads one ending in an underscore as a string.
var yamlTyped = regexp.MustCompile(`^(?:true|True|TRUE|false|False|FALSE|[-+]?(?:[0-9][0-9_]*(?:\.[0-9_]*)?|\.[0-9_]+)(?:[eE][-+]?[0-9]+)?|[-+]?0(?:b[01_]+|o[0-7_]+|x[0-9a-fA-F_]+)|[-+]?\.(?:inf|Inf|INF)|\.(?:nan|NaN|NAN))$`)

// keptNumber reports whether JavaScript writes the number s back as s, as
// Mermaid draws a numeric title: in its shortest form, without an exponent,
// which it takes below 1e-6 and from 1e21, and not 0, which draws no title.
func keptNumber(s string) bool {
	v, err := strconv.ParseFloat(s, 64)
	a := math.Abs(v)
	return err == nil && a >= 1e-6 && a < 1e21 && strconv.FormatFloat(v, 'f', -1, 64) == s
}

// yamlScalar reads v, a value on one line, as YAML reads a string: a plain
// value ends at a comment, a single-quoted one doubles its quotes and a
// double-quoted one takes the escapes \" \\ \n \t and \uXXXX; only a comment
// may follow the closing quote. ok is false for anything else YAML could
// read there, which would not be the string written: a null, a block
// scalar, a collection, an anchor, an unclosed quote, a bool, or a number
// Mermaid would not draw as written (0 draws no title, 1.50 draws 1.5).
func yamlScalar(v string) (s string, ok bool) {
	v = strings.TrimSpace(v)
	if v == "" || v[0] != '"' && v[0] != '\'' {
		// A comment starts at a # after a space, or in place of the value.
		v, _, _ = strings.Cut(" "+strings.ReplaceAll(v, "\t", " "), " #")
		v = strings.TrimSpace(v)
		switch {
		case v == "":
			return "", true
		case v == "~", v == "null", v == "Null", v == "NULL",
			strings.ContainsRune("[]{},&*!|>%@`", rune(v[0])),
			strings.HasPrefix(v+" ", "- "), strings.HasPrefix(v+" ", "? "),
			strings.Contains(v+" ", ": "),
			yamlTyped.MatchString(v) && !strings.HasSuffix(v, "_") && !keptNumber(v):
			return "", false
		}
		return v, true
	}
	q := v[0]
	var b strings.Builder
	for i := 1; i < len(v); i++ {
		switch c := v[i]; {
		case c == '\'' && q == '\'' && strings.HasPrefix(v[i+1:], "'"):
			b.WriteByte(c)
			i++
		case c == q:
			rest := v[i+1:]
			if after := strings.TrimLeft(rest, " \t"); after != "" && (after == rest || after[0] != '#') {
				return "", false
			}
			return b.String(), true
		case c == '\\' && q == '"':
			// Of the escapes UnquoteChar reads, only YAML's \" \\ \n \t \u.
			if i+1 == len(v) || !strings.ContainsRune(`"\ntu`, rune(v[i+1])) {
				return "", false
			}
			r, _, tail, err := strconv.UnquoteChar(v[i:], q)
			if err != nil {
				return "", false
			}
			b.WriteRune(r)
			i = len(v) - len(tail) - 1
		default:
			b.WriteByte(c)
		}
	}
	return "", false // unclosed
}

// continues reports whether line goes on with the value of the key above it:
// it is indented, and more than a comment.
func continues(line string) bool {
	text := strings.TrimSpace(line)
	return text != "" && text[0] != '#' && (line[0] == ' ' || line[0] == '\t')
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

// equalFold reports whether s and t are equal ignoring ASCII case: every
// caller passes an ASCII keyword, and a non-ASCII rune that folds to ASCII
// takes two bytes or more, so equal lengths keep the match ASCII-only.
func equalFold(s, t string) bool { return len(s) == len(t) && strings.EqualFold(s, t) }

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
// `title: Text`, and returns the title as drawn (see plain).
func title(text string) (string, bool) {
	if v, ok := statement(text, "title"); ok {
		return plain(v), true
	}
	if rest, ok := cutPrefixFold(text, "title "); ok {
		return plain(rest), true
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
