package mathml

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMMLNodeWriteEscapes(t *testing.T) {
	n := NewMMLNode("mi", `</math><script>x & y</script>`)
	n.SetAttr("class", `x" onclick="alert(1)`)
	n.SetAttr(`bad" onclick="alert(1)`, "v")
	var b strings.Builder
	n.Write(&b)
	assert.Equal(t,
		`<mi class="x&#34; onclick=&#34;alert(1)">&lt;/math&gt;&lt;script&gt;x &amp; y&lt;/script&gt;</mi>`,
		b.String())
}

// Node text holds literal characters: an entity reference typed by the
// author must display as typed, not be decoded by the browser.
func TestMMLNodeWriteEscapesEntities(t *testing.T) {
	for in, want := range map[string]string{
		"&OverBrace;":    "&amp;OverBrace;",
		"&lt;b&gt;":      "&amp;lt;b&amp;gt;",
		"&#34;":          "&amp;#34;",
		"&#x2061;":       "&amp;#x2061;",
		"&notanentity &": "&amp;notanentity &amp;",
	} {
		var b strings.Builder
		NewMMLNode("mo", in).Write(&b)
		assert.Equal(t, "<mo>"+want+"</mo>", b.String())
	}
}

func TestConvertKeepsTypedEntities(t *testing.T) {
	out, err := NewMathMLConverter().ConvertInline(`\text{&lt;b&gt; a b} \overbrace{x} < y`)
	assert.NoError(t, err)
	assert.Contains(t, out, "<mtext>&amp;lt;b&amp;gt;\u00a0a\u00a0b</mtext>")
	assert.Contains(t, out, "<mo stretchy=\"true\">⏞</mo>")
	assert.Contains(t, out, "<mo>&lt;</mo>")
	assert.Contains(t, out, `<annotation encoding="application/x-tex">\text{&amp;lt;b&amp;gt; a b}`)
}

func TestParseDepthLimit(t *testing.T) {
	deep := strings.Repeat(`\sqrt{`, MaxParseDepth+1) + "x" + strings.Repeat("}", MaxParseDepth+1)
	_, err := NewMathMLConverter().ConvertInline(deep)
	assert.Error(t, err)

	shallow := strings.Repeat(`\sqrt{`, 10) + "x" + strings.Repeat("}", 10)
	_, err = NewMathMLConverter().ConvertInline(shallow)
	assert.NoError(t, err)
}

func TestMMLNodeWriteSortsAttributes(t *testing.T) {
	n := NewMMLNode("mo", "lim").SetAttr("rspace", "0").SetAttr("fence", "true").SetAttr("lspace", "0")
	for range 20 {
		var b strings.Builder
		n.Write(&b)
		assert.Equal(t, `<mo fence="true" lspace="0" rspace="0">lim</mo>`, b.String())
	}
}

// The page CSP blocks inline styles: the converter describes the math and
// the stylesheet styles it, through classes.
func TestNoInlineStyle(t *testing.T) {
	for tex, class := range map[string]string{
		`x`:                 "",
		`\dot{\imath}`:      `class="math-dtls-on"`,
		`\ddot{x}`:          `class="math-dtls-on"`,
		`\underline{x}`:     `class="math-dtls-on"`,
		`\LaTeX`:            `class="math-latex-a"`,
		`\TeX`:              `class="math-tex-e"`,
		`\varliminf_n x`:    `class="math-liminf"`,
		`\varlimsup_n x`:    `class="math-limsup"`,
		`\dot{\mathcal{A}}`: `class="mathcal math-dtls-on"`,
	} {
		for _, display := range []bool{false, true} {
			out := convertWithin(t, tex, display)
			assert.NotContains(t, out, " style=", tex)
			assert.Contains(t, out, class, tex)
		}
	}
}

// \raisebox shifts content without moving the box around it, so a large
// shift would let math draw over the page around it.
func TestRaiseboxShiftIsBounded(t *testing.T) {
	for arg, want := range map[string]string{
		"1em":             `voffset="1em"`,
		"-0.5ex":          `voffset="-0.5ex"`,
		"2pt":             `voffset="2pt"`,
		"1":               `voffset="1em"`,
		"+1em":            `voffset="1em"`,
		" 0.3 em ":        `voffset="0.3em"`,
		"1EM":             `voffset="1em"`,
		"0.5cm":           `voffset="0.5cm"`,
		"-3mm":            `voffset="-3mm"`,
		"0.2in":           `voffset="0.2in"`,
		"1pc":             `voffset="1pc"`,
		"10px":            `voffset="10px"`,
		"3mu":             `voffset="0.1667em"`,
		"-18mu":           `voffset="-1em"`,
		"10bp":            `voffset="1.0037em"`,
		"2dd":             `voffset="0.214em"`,
		"1cc":             `voffset="1.284em"`,
		"65536sp":         `voffset="0.1em"`,
		"-1000em":         "",
		"2.1em":           "",
		"1cm":             "",
		"1in":             "",
		"2pc":             "",
		"99999px":         "",
		"37mu":            "",
		"1em;x":           "",
		"0.5em;color:red": "",
		"calc(1em)":       "",
		"1e3":             "",
		"1e9em":           "",
		"":                "",
		"em":              "",
		"1fil":            "",
		"1xx":             "",
		"--1em":           "",
	} {
		out, err := NewMathMLConverter().ConvertInline(`\raisebox{` + arg + `}{x}`)
		assert.NoError(t, err)
		if want == "" {
			assert.NotContains(t, out, "voffset", arg)
		} else {
			assert.Contains(t, out, want, arg)
		}
		assert.Contains(t, out, "<mi>x</mi>", arg)
	}
}

func TestCellSpanIsBounded(t *testing.T) {
	for arg, want := range map[string]string{
		"2":     `rowspan="2"`,
		" 3 ":   `rowspan="3"`,
		"64":    `rowspan="64"`,
		"65":    "",
		"30000": "",
		"0":     "",
		"-1":    "",
		"2x":    "",
	} {
		out, err := NewMathMLConverter().ConvertDisplay(`\begin{array}{cc}\multirow{` + arg + `}{*}{\uparrow} & a \\ & b\end{array}`)
		assert.NoError(t, err)
		if want == "" {
			assert.NotContains(t, out, "rowspan", arg)
			assert.NotContains(t, out, "minsize", arg)
		} else {
			assert.Contains(t, out, want, arg)
		}
	}
}

// A \multicolumn cell is one cell of the source: the cell after it starts
// after the columns it covers and must not be dropped, even when the span
// already fills the row. A cell typed where a \multirow cell above leaves
// its place is not dropped either.
func TestTableCellsAfterSpansAreKept(t *testing.T) {
	for _, c := range []struct{ name, tex string }{
		{"multicolumn fills the row", `\begin{array}{cc}\multicolumn{2}{c}{e}&f\end{array}`},
		{"multicolumn leaves a column", `\begin{array}{ccc}\multicolumn{2}{c}{e}&f\end{array}`},
		{"multicolumn in a later row", `\begin{array}{ccc}a&b&c\\ \multicolumn{2}{c}{e}&f\end{array}`},
		{"text under multirow", `\begin{array}{cc}\multirow{2}{*}{e} & a \\ f & b\end{array}`},
	} {
		t.Run(c.name, func(t *testing.T) {
			out, err := NewMathMLConverter().ConvertDisplay(c.tex)
			require.NoError(t, err)
			assert.Contains(t, out, "<mi>e</mi>")
			assert.Contains(t, out, "<mi>f</mi>")
		})
	}
}

// A cell left empty under a \multirow cell gets no <mtd>, and the cells
// after it keep their column's alignment.
func TestTableColumnsAfterSpans(t *testing.T) {
	out, err := NewMathMLConverter().ConvertDisplay(`\begin{array}{lcr}\multicolumn{2}{c}{e}&f\\ \multirow{2}{*}{g}&h&i\\ &j&k\end{array}`)
	require.NoError(t, err)
	assert.Equal(t, 7, strings.Count(out, "<mtd"))
	assert.Contains(t, out, "<mtd columnalign=\"right\"><mi>f</mi>")
	assert.Contains(t, out, "<mtd columnalign=\"center\"><mi>j</mi>")
	assert.Contains(t, out, "<mtd columnalign=\"right\"><mi>k</mi>")
}

// convertWithin converts tex, failing the test if conversion does not
// finish within two seconds instead of hanging the test binary.
func convertWithin(t *testing.T, tex string, display bool) string {
	t.Helper()
	done := make(chan string, 1)
	go func() {
		c := NewMathMLConverter()
		var out string
		if display {
			out, _ = c.ConvertDisplay(tex)
		} else {
			out, _ = c.ConvertInline(tex)
		}
		done <- out
	}()
	select {
	case out := <-done:
		return out
	case <-time.After(2 * time.Second):
		t.Fatalf("conversion of %q did not terminate", tex)
		return ""
	}
}

// Every loop reading a token buffer must advance it, or the input hangs the
// renderer. A brace group where a token is expected is the usual trap.
func TestConversionTerminates(t *testing.T) {
	for _, tex := range []string{
		`\sideset{{a}}{}{x}`,
		`\sideset{}{{a}}{x}`,
		`\sideset{{a}^b}{_{c}{d}}{\sum}`,
		`\sideset{{}}{{}}{x}`,
		`\sideset{ {a} }{ {b} }{x}`,
		`\frac{{a}}{{b}}`,
		`\frac {a} {b}`,
		`\frac`,
		`\sqrt[{3}]{{x}}`,
		`\textcolor{{red}}{x}`,
		`\newcommand{{\x}}{y}`,
		`\color{{red}} x`,
		`\begin{array}{{c}}a\end{array}`,
		`\left\{ {a} \right.`,
		`x^{} _{} {}{}{}`,
		`\mathbf{{x}}`,
		`\hat{{x}}`,
		`\raisebox{{1em}}{{x}}`,
		`\multirow{{2}}{*}{{x}}`,
		`\prescript{{a}}{{b}}{{c}}`,
	} {
		convertWithin(t, tex, false)
		convertWithin(t, tex, true)
	}
}

// FuzzConversionTerminates checks that no input makes the converter loop.
func FuzzConversionTerminates(f *testing.F) {
	for _, s := range []string{
		`\sideset{_a^b}{_c}{\sum}`, `\frac{a}{b}`, `\begin{array}{cc}a&b\\c&d\end{array}`,
		`\left( x \middle| y \right)`, `\color{red} x`, `\newcommand{\x}{y}`,
		`\sqrt[3]{x}`, `\raisebox{1em}{x}`, `\bigl( x \bigr)`, `a \over b`,
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, tex string) {
		if len(tex) > 512 {
			return
		}
		convertWithin(t, tex, false)
	})
}

// \color and \textcolor accept theme colours only, as classes the
// stylesheet colours; any other colour leaves the content uncoloured.
func TestColorIsThemeClass(t *testing.T) {
	for arg, want := range map[string]string{
		"red":        "math-color-red",
		"Blue":       "math-color-blue",
		" green ":    "math-color-green",
		"ORANGE":     "math-color-orange",
		"purple":     "math-color-purple",
		"gray":       "math-color-gray",
		"grey":       "math-color-gray",
		"white":      "",
		"black":      "",
		"#FF0000":    "",
		"#fff":       "",
		"rgb(1,0,0)": "",
		"red;x":      "",
		"":           "",
	} {
		for _, tex := range []string{`\color{` + arg + `} x`, `\textcolor{` + arg + `}{x}`} {
			out := convertWithin(t, tex, false)
			assert.NotContains(t, out, "mathcolor", tex)
			assert.Contains(t, out, "<mi>x</mi>", tex)
			if want == "" {
				assert.NotContains(t, out, "math-color-", tex)
			} else {
				assert.Contains(t, out, `<mstyle class="`+want+`">`, tex)
			}
		}
	}
	// Empty content does not fail the conversion.
	_, err := NewMathMLConverter().ConvertInline(`\textcolor{red}{}`)
	assert.NoError(t, err)
}

// A mathsize percentage is relative to the enclosing size, so nested size
// switches multiply; the cumulative size stays within \tiny and \Huge.
func TestNestedSizeIsBounded(t *testing.T) {
	sizeRe := regexp.MustCompile(`mathsize="([0-9.]+)%"`)
	for tex, want := range map[string][]string{
		`\Huge x`:                     {"248.8"},
		`\tiny x`:                     {"050.0"},
		`\Huge \Huge x`:               {"248.8", "100.0"},
		`\Huge{\Huge{\Huge x}}`:       {"248.8", "100.0", "100.0"},
		`\large \Huge x`:              {"120.0", "207.3"},
		`\Huge \tiny x`:               {"248.8", "050.0"},
		`\tiny \tiny x`:               {"050.0", "100.0"},
		`\tiny \Huge x`:               {"050.0", "248.8"},
		`\scriptsize \footnotesize x`: {"070.0", "080.0"},
		`{\Huge a} {\Huge b}`:         {"248.8", "248.8"},
	} {
		out := convertWithin(t, tex, false)
		var got []string
		for _, m := range sizeRe.FindAllStringSubmatch(out, -1) {
			got = append(got, m[1])
		}
		assert.Equal(t, want, got, tex)
	}
	deep := strings.Repeat(`\Huge{`, 40) + "x" + strings.Repeat("}", 40)
	out := convertWithin(t, deep, false)
	scale := 1.0
	for _, m := range sizeRe.FindAllStringSubmatch(out, -1) {
		v, err := strconv.ParseFloat(m[1], 64)
		require.NoError(t, err)
		scale *= v / 100
	}
	assert.LessOrEqual(t, scale, 2.4881)
}

// Nested \raisebox shifts add up: their sum stays within the 2em a single
// \raisebox may shift. A shift that would exceed it is dropped, and its
// content still rendered.
func TestNestedRaiseIsBounded(t *testing.T) {
	voffRe := regexp.MustCompile(`voffset="([^"]*)"`)
	for tex, want := range map[string][]string{
		`\raisebox{2em}{x}`:                                     {"2em"},
		`\raisebox{1em}{\raisebox{1em}{x}}`:                     {"1em", "1em"},
		`\raisebox{2em}{\raisebox{2em}{x}}`:                     {"2em"},
		`\raisebox{2em}{\raisebox{-2em}{x}}`:                    {"2em", "-2em"},
		`\raisebox{1.5em}{\raisebox{1em}{\raisebox{0.5em}{x}}}`: {"1.5em", "0.5em"},
		`\raisebox{2em}{x}\raisebox{2em}{y}`:                    {"2em", "2em"},
		`\raisebox{10pt}{\raisebox{1em}{\raisebox{1pt}{x}}}`:    {"10pt", "1em"},
	} {
		out := convertWithin(t, tex, false)
		var got []string
		for _, m := range voffRe.FindAllStringSubmatch(out, -1) {
			got = append(got, m[1])
		}
		assert.Equal(t, want, got, tex)
		assert.Contains(t, out, "<mi>x</mi>", tex)
	}
	deep := strings.Repeat(`\raisebox{1em}{`, 40) + "x" + strings.Repeat("}", 40)
	out := convertWithin(t, deep, false)
	assert.Equal(t, 2, strings.Count(out, "voffset="))
}

// Inline math is written without surrounding or inner whitespace, which
// would show as spaces around it, as in "($x$)".
func TestInlineOutputIsCompact(t *testing.T) {
	out := convertWithin(t, `\frac{a}{b} + \begin{pmatrix}1&2\end{pmatrix}`, false)
	assert.NotContains(t, out, "\n")
	assert.True(t, strings.HasPrefix(out, "<math "), out)
	assert.True(t, strings.HasSuffix(out, "</math>"), out)
	assert.NotContains(t, out, "> ")
	out = convertWithin(t, `\frac{a}{b}`, true)
	assert.Equal(t, 1, strings.Count(out, "\n"))
	assert.True(t, strings.HasSuffix(out, "</math>\n"), out)
}

func BenchmarkConvert(b *testing.B) {
	formulas := []string{
		`x^2 + y^2 = z^2`,
		`\frac{-b \pm \sqrt{b^2 - 4ac}}{2a}`,
		`\sum_{n=1}^{\infty} \frac{1}{n^2} = \frac{\pi^2}{6}`,
		`\int_{-\infty}^{\infty} e^{-x^2}\,dx = \sqrt{\pi}`,
		`\begin{pmatrix} a & b \\ c & d \end{pmatrix}\begin{pmatrix} x \\ y \end{pmatrix}`,
		`\mathbb{E}[X] = \sum_{i} x_i \, \Pr(X = x_i)`,
		`\left( \sum_{k=1}^n a_k b_k \right)^2 \le \left( \sum_{k=1}^n a_k^2 \right) \left( \sum_{k=1}^n b_k^2 \right)`,
		`\begin{aligned} f(x) &= (x+1)^2 \\ &= x^2 + 2x + 1 \end{aligned}`,
	}
	b.ReportAllocs()
	for b.Loop() {
		for _, f := range formulas {
			if _, err := NewMathMLConverter().ConvertDisplay(f); err != nil {
				b.Fatal(err)
			}
		}
	}
}
