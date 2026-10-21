package test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// `// GnoTrim: lines 12, 40-45` is a hand-written tail directive of a
// corpus file: the harness blanks the listed body lines before parsing,
// so a file skipped for an incidental unsupported construct (an
// `import "unsafe"` and its few uses, a stray complex or imaginary
// literal, a `uintptr`) is still checked. Line numbers do not move, so
// markers keep their lines; a trimmed line that carried a marker is
// excluded from the expected set and listed as `// GnoTrimmed: lines …`.
// Both are golden-region lines: sync keeps the directive verbatim and
// rewrites the evidence; verify refuses a directive that names a line
// outside the body or one without a trimmable construct.
const (
	DirectiveGnoTrim    = "GnoTrim"
	DirectiveGnoTrimmed = "GnoTrimmed"
)

// trimmableConstructs are the only things a GnoTrim line may remove: the
// four incidental feature gaps (unsupportedTags unsafe, complex-types,
// imaginary-literal, uintptr). A listed line must match one of them.
var trimmableConstructs = []struct {
	tag string
	re  *regexp.Regexp
}{
	{"unsafe", regexp.MustCompile(`"unsafe"|\bunsafe\.`)},
	{"complex-types", regexp.MustCompile(`\bcomplex(64|128)?\b|\b(real|imag)\(`)},
	{"imaginary-literal", regexp.MustCompile(`\b\d[\w.]*i\b|\.\d+i\b`)},
	{"uintptr", regexp.MustCompile(`\buintptr\b`)},
}

// gnoTrim is a parsed `// GnoTrim:` directive.
type gnoTrim struct {
	text  string // as written, kept verbatim through sync
	lines []int  // ascending, unique
}

var reTrimItem = regexp.MustCompile(`^(\d+)(?:-(\d+))?$`)

// parseGnoTrim reads the content of a `// GnoTrim:` line:
// `lines N[-M], N[-M], …`, 1-based, ranges inclusive, no line twice.
func parseGnoTrim(content string) (gnoTrim, error) {
	content = strings.TrimSpace(content)
	rest, ok := strings.CutPrefix(content, "lines ")
	if !ok {
		return gnoTrim{}, fmt.Errorf("`// GnoTrim:` wants `lines N, N-M, …`, got %q", content)
	}
	t := gnoTrim{text: content}
	seen := map[int]bool{}
	for item := range strings.SplitSeq(rest, ",") {
		m := reTrimItem.FindStringSubmatch(strings.TrimSpace(item))
		if m == nil {
			return gnoTrim{}, fmt.Errorf("`// GnoTrim:` wants `lines N, N-M, …`, got %q", content)
		}
		lo, _ := strconv.Atoi(m[1])
		hi := lo
		if m[2] != "" {
			hi, _ = strconv.Atoi(m[2])
		}
		if lo < 1 || hi < lo {
			return gnoTrim{}, fmt.Errorf("`// GnoTrim:` bad range %q", strings.TrimSpace(item))
		}
		for n := lo; n <= hi; n++ {
			if seen[n] {
				return gnoTrim{}, fmt.Errorf("`// GnoTrim:` lists line %d twice", n)
			}
			seen[n] = true
			t.lines = append(t.lines, n)
		}
	}
	return t, nil
}

// apply returns body with the directive's lines blanked, each line checked
// to exist and to carry a trimmable construct. Blanking every marker line
// would turn an errorcheck file into a compile file, so that is refused.
func (t gnoTrim) apply(body []byte) ([]byte, error) {
	src := string(body)
	lines := strings.Split(src, "\n")
	n := len(lines)
	if lines[n-1] == "" { // trailing newline
		n--
	}
	for _, l := range t.lines {
		if l > n {
			return nil, fmt.Errorf("`// GnoTrim:` names line %d, but the body has %d lines", l, n)
		}
		if trimTag(lines[l-1]) == "" {
			return nil, fmt.Errorf("`// GnoTrim:` line %d carries no unsafe, complex, imaginary or uintptr construct: %q", l, lines[l-1])
		}
		lines[l-1] = ""
	}
	out := []byte(strings.Join(lines, "\n"))
	if len(ParseInlineErrors(body, false)) > 0 && len(ParseInlineErrors(out, false)) == 0 {
		return nil, fmt.Errorf("`// GnoTrim:` blanks every `// ERROR` marker line")
	}
	return out, nil
}

// trimTag names the trimmable construct on line, or "".
func trimTag(line string) string {
	for _, c := range trimmableConstructs {
		if c.re.MatchString(line) {
			return c.tag
		}
	}
	return ""
}

// markedLines lists the directive's lines that carry a marker in body:
// gc's expectations the trim drops, reported as `// GnoTrimmed:`.
func (t gnoTrim) markedLines(body []byte) []int {
	trimmed := map[int]bool{}
	for _, l := range t.lines {
		trimmed[l] = true
	}
	var out []int
	for _, mk := range ParseInlineErrors(body, false) {
		if trimmed[mk.Line] {
			out = append(out, mk.Line)
		}
	}
	return out
}

// trimSections are the two golden-region lines a trim adds: the directive
// as written and the marker lines it dropped; both empty without a trim,
// so a stray one is still caught.
func trimSections(trim *gnoTrim, body []byte) []goldenSection {
	s := []goldenSection{
		{name: DirectiveGnoTrim, inline: true},
		{name: DirectiveGnoTrimmed, inline: true},
	}
	if trim != nil {
		s[0].block = trim.text
		if ls := trim.markedLines(body); len(ls) > 0 {
			s[1].block = "lines " + lineList(ls)
		}
	}
	return s
}

// lineList renders ns as "3, 9".
func lineList(ns []int) string {
	ss := make([]string, len(ns))
	for i, n := range ns {
		ss[i] = strconv.Itoa(n)
	}
	return strings.Join(ss, ", ")
}
