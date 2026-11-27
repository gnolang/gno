package mermaid

import (
	"bytes"
	"encoding/xml"
	"errors"
	"flag"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update-golden-tests", false, "update golden tests")

// TestGolden draws each testdata/*.mmd and compares the SVG with the .svg
// beside it; -update-golden-tests rewrites them.
func TestGolden(t *testing.T) {
	sources, err := filepath.Glob("testdata/*.mmd")
	require.NoError(t, err)
	for _, src := range sources {
		name := strings.TrimSuffix(filepath.Base(src), ".mmd")
		t.Run(name, func(t *testing.T) {
			in, err := os.ReadFile(src)
			require.NoError(t, err)
			out, _, err := Render(string(in))
			require.NoError(t, err)
			checkSVG(t, out)
			golden := strings.TrimSuffix(src, ".mmd") + ".svg"
			if *update {
				require.NoError(t, os.WriteFile(golden, append(out, '\n'), 0o644))
				return
			}
			want, err := os.ReadFile(golden)
			require.NoError(t, err)
			assert.Equal(t, string(bytes.TrimSuffix(want, []byte("\n"))), string(out))
		})
	}
}

func TestTabs(t *testing.T) {
	// Mermaid separates words with tabs as with spaces.
	for _, src := range []string{
		"pie\tshowData\n\"a\" : 1",
		"sequenceDiagram\nparticipant\tA\nNote\tright of\tA: x",
		"xychart-beta\nx-axis\t[a]\nbar\t[1]",
		"timeline\n2020 :\ta\nsection\tS\n2021 : b",
	} {
		_, _, err := Render(src)
		assert.NoError(t, err, src)
	}
}

// renderErrors checks that each source fails with an *Error whose message
// holds the given text.
func renderErrors(t *testing.T, cases map[string][2]string) {
	t.Helper()
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			out, _, err := Render(c[0])
			var e *Error
			require.True(t, errors.As(err, &e), "got %v", err)
			assert.Contains(t, e.Error(), c[1])
			assert.Nil(t, out)
		})
	}
}

func TestRenderErrors(t *testing.T) {
	renderErrors(t, map[string][2]string{
		"empty":              {"", "empty diagram"},
		"only comments":      {"%% a comment\n%%{init: {}}%%\n", "empty diagram"},
		"unknown type":       {"graphviz\n  a -> b", `unsupported diagram type "graphviz"`},
		"too long":           {"pie\n" + strings.Repeat(" ", MaxSourceLen), "longer than"},
		"invalid utf-8":      {"pie\n\"a\xff\" : 1", "UTF-8"},
		"unclosed front":     {"---\ntitle: x\npie", "frontmatter"},
		"unclosed directive": {"%%{init: {\n\"theme\": \"dark\"\npie\n\"a\" : 1", "directive"},
	})
}

func TestScan(t *testing.T) {
	var m meta
	lines, err := scan("\n---\ntitle: \"From front\"\nconfig:\n  title: nested\n---\n%%{init: {\n  \"theme\": \"dark\"\n}}%%\npie\n  %% a comment\n\t\"50%% off\" : 1\naccDescr { Our share }\n", &m)
	require.NoError(t, err)
	assert.Equal(t, "From front", m.title, "a title nested under config is not the diagram's")
	assert.Equal(t, []line{{n: 10, text: "pie"}, {n: 12, indent: 1, text: `"50%% off" : 1`}, {n: 13, text: "accDescr { Our share }"}}, lines,
		"only whole-line comments are dropped, as Mermaid does")
	body, err := accessibility(lines[1:], &m)
	require.NoError(t, err)
	assert.Equal(t, "Our share", m.accDescr)
	assert.Equal(t, lines[1:2], body)
}

func TestFrontmatterTitle(t *testing.T) {
	// The title is a YAML scalar on one line.
	for value, want := range map[string]string{
		`My chart # draft`:    "My chart",
		`C# notes`:            "C# notes",
		`'It''s'`:             "It's",
		`"a \"b\" \\ \u00e9"`: `a "b" \ é`,
		`"'kept'"  # quoted`:  "'kept'",
		`# only a comment`:    "",
		``:                    "",
	} {
		var m meta
		_, err := scan("---\ntitle: "+value+"\n---\npie", &m)
		require.NoError(t, err, value)
		assert.Equal(t, want, m.title, value)
	}
	var m meta
	_, err := scan("---\nTitle: x\n---\npie", &m)
	require.NoError(t, err)
	assert.Empty(t, m.title, "YAML keys are case-sensitive")

	for _, value := range []string{
		">", "|-", "~", "null", "'unclosed", `"unclosed`, `"x" y`, `'x'#c`, `"\q"`, `"\ud800"`,
		"[a, b]", "&anchor x", "*alias", "- x", "Time: 10:30", "a\n  continued", "\n  next line",
	} {
		_, err := scan("---\ntitle: "+value+"\n---\npie", &meta{})
		var e *Error
		assert.True(t, errors.As(err, &e), "%q: got %v", value, err)
	}

	// Render draws the title as YAML reads it, without unquoting it again.
	out, _, err := Render("---\ntitle: \"'Q' #59;\"\n---\npie\n\"a\" : 1")
	require.NoError(t, err)
	assert.Contains(t, string(out), ">&#39;Q&#39; #59;<")
}

// toLines makes the lines of a diagram's body, numbered from 1.
func toLines(texts ...string) []line {
	ls := make([]line, len(texts))
	for i, s := range texts {
		ls[i] = line{n: i + 1, text: s}
	}
	return ls
}

func TestTitle(t *testing.T) {
	for in, want := range map[string]string{"title x y": "x y", "Title: x": "x", "TITLE  x": "x"} {
		got, ok := title(in)
		assert.True(t, ok, in)
		assert.Equal(t, want, got, in)
	}
	_, ok := title("titles")
	assert.False(t, ok)
}

func TestLabel(t *testing.T) {
	for in, want := range map[string][]string{
		"one":                     {"one"},
		"a<br>b":                  {"a", "b"},
		"a <BR/> b<br />c":        {"a", "b", "c"},
		"<brand> and <b>bold</b>": {"<brand> and <b>bold</b>"},
		"end<br":                  {"end<br"},
		"<br>":                    {"", ""},
		"#60;br#62;":              {"<br>"}, // an encoded tag stays text
	} {
		assert.Equal(t, want, label(in), in)
	}
}

func TestDecodeEntities(t *testing.T) {
	for in, want := range map[string]string{
		"a#59;b":         "a;b",
		"#quot;q#quot;":  `"q"`,
		"#9829;":         "♥",
		"Fix #42 now":    "Fix #42 now",
		"#unknown;":      "#unknown;",
		"#0;#7;#55296;":  "#0;#7;#55296;", // controls and surrogates stay
		"#99999999999;":  "#99999999999;",
		"##59;":          "#;",
		"#":              "#",
		"#lt;script#gt;": "<script>",
	} {
		assert.Equal(t, want, decodeEntities(in), in)
	}
}

func TestTextWidth(t *testing.T) {
	// "Hi" is H plus i, in thousandths of an em, at 1000 px.
	h, i := interWidths['H'], interWidths['i']
	assert.Equal(t, float64(h[0]+i[0]), textWidth("Hi", 1000, false))
	assert.Equal(t, float64(h[1]+i[1]), textWidth("Hi", 1000, true))
	assert.Equal(t, 1000.0, textWidth("漢", 1000, false), "East Asian falls back to a full em")
	assert.Equal(t, 600.0, textWidth("Ж", 1000, false), "other scripts fall back to 0.6 em")
	assert.Equal(t, 0.0, textWidth("́", 1000, false), "a combining mark takes no room")
}

func TestWrap(t *testing.T) {
	narrow := textWidth("word", fontSize, false) + 1
	assert.Equal(t, []string{"word", "word", "word"}, wrap([]string{"word word word"}, fontSize, narrow, false))
	assert.Equal(t, []string{"a b", "c"}, wrap([]string{"a b", "c"}, fontSize, 1000, false))
	assert.Equal(t, []string{""}, wrap([]string{""}, fontSize, 1000, false))

	// A word wider than the line breaks after a separator when it can, and
	// between characters otherwise; no piece is wider than the line.
	width := textWidth("gno.land/r/gov/", fontSize, false) + 1
	got := wrap([]string{"see gno.land/r/gov/dao/v3/impl/memberstore"}, fontSize, width, false)
	assert.Equal(t, "see", got[0])
	assert.Equal(t, "gno.land/r/gov/", got[1])
	assert.Equal(t, "see gno.land/r/gov/dao/v3/impl/memberstore", strings.Replace(strings.Join(got, ""), "see", "see ", 1))
	for _, l := range append(got, wrap([]string{"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", "漢字漢字漢字漢字"}, fontSize, 60, false)...) {
		assert.LessOrEqual(t, textWidth(l, fontSize, false), max(width, 60.0), l)
	}

	// A no-break space holds its words together.
	nbsp := textWidth("word\u00a0word", fontSize, false) + 1
	assert.Equal(t, []string{"word word", "word\u00a0word"}, wrap([]string{"word word word\u00a0word"}, fontSize, nbsp, false))

	// No piece is wider than the line, even when a break after a separator
	// leaves a wide character to carry.
	w60 := textWidth("ab.", fontSize, false) + 1
	for _, p := range breakWord("ab.WWWWWWWW", fontSize, w60, false) {
		assert.LessOrEqual(t, textWidth(p, fontSize, false), w60, p)
	}

	// A piece never starts with a combining mark, a skin tone or a Hangul
	// vowel or final, nor splits an emoji sequence, after a wide character
	// included.
	for _, word := range []string{
		strings.Repeat("e\u0301", 40),
		strings.Repeat("漢\u0301", 40),
		strings.Repeat("👩\u200d👩\u200d👧\ufe0f", 20),
		strings.Repeat("👍🏽", 6),
		strings.Repeat("कि", 40),
		strings.Repeat("\u1100\u1161\u11a8", 20),
	} {
		for _, width := range []float64{50, 60} {
			pieces := breakWord(word, fontSize, width, false)
			require.Greater(t, len(pieces), 1, word)
			for _, p := range pieces[1:] {
				r, _ := utf8.DecodeRuneInString(p)
				assert.False(t, unicode.In(r, unicode.Mn, unicode.Me, unicode.Mc, unicode.Cf), "%q starts with %U", p, r)
				assert.False(t, r == 0x1f3fd || r == 0x1161 || r == 0x11a8, "%q starts with %U", p, r)
			}
			for _, p := range pieces {
				assert.False(t, strings.HasSuffix(p, "\u200d"), "%q ends with a joiner", p)
			}
			assert.Equal(t, word, strings.Join(pieces, ""))
		}
	}

	// Flags break between them, never inside one.
	for _, width := range []float64{30, 50} {
		pieces := breakWord(strings.Repeat("🇫🇷", 12), fontSize, width, false)
		require.Greater(t, len(pieces), 1)
		for _, p := range pieces {
			assert.Equal(t, strings.Repeat("🇫🇷", utf8.RuneCountInString(p)/2), p)
		}
	}
}

func TestWriterEscapes(t *testing.T) {
	var w svgWriter
	w.text(classMuted, 1, 2, anchorStart, 0, `<script>alert("x")</script>`, "a & b\x00")
	out := w.document(10, 10, 10, `<title>`, `"><img src=x onerror=alert(1)>`, "</desc><script>")
	checkSVG(t, out)
	s := string(out)
	assert.NotContains(t, s, "<script")
	assert.NotContains(t, s, "<img")
	assert.Contains(t, s, "&lt;script&gt;")
	assert.Contains(t, s, "a &amp; b�")
}

func TestWriterBound(t *testing.T) {
	// Past MaxOutputLen a writer drops what it is given, so a drawing Render
	// refuses is not drawn in full first.
	var w svgWriter
	for w.b.Len() <= MaxOutputLen {
		w.rect(classBox, 1, 2, 3, 4, 5)
	}
	n := w.b.Len()
	w.rect(classBox, 1, 2, 3, 4, 5)
	w.text(classMuted, 1, 2, anchorMiddle, 0, "a & b", "c")
	w.path(classStroke, new(pathData).cmd('M', 1, 2))
	assert.Equal(t, n, w.b.Len())
}

func TestWriteNum(t *testing.T) {
	for v, want := range map[float64]string{
		0: "0", -0.001: "0", 1.005: "1", 2.5: "2.5", 1.234: "1.23", -3: "-3",
		1e307: "1000000000", -1e307: "-1000000000", math.Inf(1): "0", math.NaN(): "0",
	} {
		var b buffer
		writeNum(&b, v)
		assert.Equal(t, want, b.String(), "%v", v)
	}
}

func TestParseNumber(t *testing.T) {
	for _, ok := range []string{"1", "-2.5", "+3", ".5", "4.", "1e3", "1E-2"} {
		_, err := parseNumber(ok)
		assert.NoError(t, err, ok)
	}
	for _, bad := range []string{"", "1_000", "0x10", "0x1p4", "Inf", "nan", "1e16", " 1", "1,5"} {
		_, err := parseNumber(bad)
		assert.Error(t, err, bad)
	}
}

// FuzzRender checks that any source either fails with an *Error or draws an
// SVG made only of what svgWriter writes, within MaxOutputLen.
func FuzzRender(f *testing.F) {
	sources, _ := filepath.Glob("testdata/*.mmd")
	for _, src := range sources {
		in, _ := os.ReadFile(src)
		f.Add(string(in))
	}
	for _, seed := range []string{
		"pie\n\"<x>\" : 1",
		"sequenceDiagram\nNote over A,: x",
		"sequenceDiagram\nautonumber 10 5\nA<<->>B: x; B-)A: #59;y",
		"timeline\n2021 : at 10:30 : https://x.y",
		"xychart-beta\nx-axis \"t [s]\" [a, b]\ny-axis 0 --> 5\nbar [-1, 9, 3]",
		"kanban\nc\n  a[\"q\"]@{ label: x, priority: 'Very Low' }",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, src string) {
		out, _, err := Render(src)
		if err != nil {
			var e *Error
			require.True(t, errors.As(err, &e), "error %v is not an *Error", err)
			require.False(t, strings.HasPrefix(e.Msg, "internal error"), "panic: %v", err)
			return
		}
		require.LessOrEqual(t, len(out), MaxOutputLen)
		checkSVG(t, out)
	})
}

// allowed lists the elements an SVG may hold and their attributes besides
// class.
var allowed = map[string][]string{
	"svg":      {"xmlns", "width", "height", "viewBox", "font-size", "role", "aria-label"},
	"desc":     nil,
	"g":        {"transform"},
	"rect":     {"x", "y", "width", "height", "rx"},
	"line":     {"x1", "y1", "x2", "y2"},
	"circle":   {"cx", "cy", "r"},
	"polygon":  {"points"},
	"polyline": {"points"},
	"path":     {"d"},
	"text":     {"x", "y", "text-anchor", "font-size"},
	"tspan":    {"x", "y"},
}

var (
	plainNumber = regexp.MustCompile(`^-?\d+(\.\d+)?$`)
	numberList  = regexp.MustCompile(`^(-?\d+(\.\d+)?[ ,]?)*$`)
	pathCmds    = regexp.MustCompile(`^([MLHVAZ]-?\d+(\.\d+)?( -?\d+(\.\d+)?)*|Z)*$`)
	translation = regexp.MustCompile(`^translate\(-?\d+(\.\d+)? -?\d+(\.\d+)?\)$`)
)

// checkSVG fails unless svg is well-formed XML made only of the elements and
// attributes in allowed, with plain numbers, path data and class names as
// values.
func checkSVG(t *testing.T, svg []byte) {
	t.Helper()
	d := xml.NewDecoder(bytes.NewReader(svg))
	depth := 0
	for {
		tok, err := d.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err, "SVG is not well-formed")
		switch tok := tok.(type) {
		case xml.StartElement:
			depth++
			name := tok.Name.Local
			attrs, ok := allowed[name]
			require.True(t, ok, "element %q is not allowed", name)
			if depth == 1 {
				require.Equal(t, "svg", name)
			}
			for _, a := range tok.Attr {
				checkAttr(t, name, a, attrs)
			}
		case xml.EndElement:
			depth--
		case xml.ProcInst, xml.Directive, xml.Comment:
			t.Fatalf("unexpected %T in SVG", tok)
		}
	}
}

func checkAttr(t *testing.T, elem string, a xml.Attr, attrs []string) {
	t.Helper()
	name, v := a.Name.Local, a.Value
	if name == "class" {
		for c := range strings.FieldsSeq(v) {
			require.True(t, strings.HasPrefix(c, "gno-mermaid__"), "class %q", c)
		}
		return
	}
	require.True(t, slices.Contains(attrs, name), "attribute %q is not allowed on <%s>", name, elem)
	switch name {
	case "xmlns":
		require.Equal(t, "http://www.w3.org/2000/svg", v)
	case "role":
		require.Equal(t, "img", v)
	case "aria-label":
		// escaped text
	case "text-anchor":
		require.Contains(t, []string{"middle", "end"}, v)
	case "d":
		require.Regexp(t, pathCmds, v)
	case "transform":
		require.Regexp(t, translation, v)
	case "points", "viewBox":
		require.Regexp(t, numberList, v)
	default:
		require.Regexp(t, plainNumber, v, "%s on <%s>", name, elem)
	}
}
