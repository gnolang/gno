package mathml

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mathBody converts tex inline and returns the MathML without its TeX
// annotation, failing the test on a conversion error.
func mathBody(t *testing.T, tex string) string {
	t.Helper()
	out, err := convertErrWithin(t, tex)
	require.NoError(t, err, tex)
	if i := strings.Index(out, "<annotation"); i >= 0 {
		out = out[:i]
	}
	return out
}

// An empty argument or group is an empty atom, as in TeX: it converts, and
// fixed-arity elements keep all their children.
func TestEmptyArguments(t *testing.T) {
	for _, tc := range []struct{ tex, want string }{
		{`\begin{array}{cc} \multicolumn{2}{c}{} \\ a & b \end{array}`, `columnspan="2"><mrow></mrow></mtd>`},
		{`\overset{a}{}`, `<mover><mrow></mrow><mi>a</mi></mover>`},
		{`\underset{a}{}`, `<munder><mrow></mrow><mi>a</mi></munder>`},
		{`\overset{}{}`, `<mover><mrow></mrow><mrow></mrow></mover>`},
		{`\substack{}`, `<mtable`},
		{`a\,{}`, `<mi>a</mi>`},
		{`T^{\mu}\,{}_{\nu}`, `<msub><mrow></mrow><mi>ν</mi></msub>`},
		{`\frac{}{b}`, `<mfrac><mrow></mrow><mi>b</mi></mfrac>`},
		{`\frac{a}{}`, `<mfrac><mi>a</mi><mrow></mrow></mfrac>`},
		{`\sqrt[]{x}`, `<msqrt><mi>x</mi></msqrt>`},
		{`\sum_{}^{n} k`, `<munderover><mo largeop="true" movablelimits="true">∑</mo><mrow></mrow><mi>n</mi></munderover>`},
		{`\int_{}^{1} f`, `<msubsup><mo largeop="true" movablelimits="true">∫</mo><mrow></mrow><mn>1</mn></msubsup>`},
		{`x^{}_2`, `<msubsup><mi>x</mi><mn>2</mn><mrow></mrow></msubsup>`},
	} {
		out := mathBody(t, tc.tex)
		assert.Contains(t, out, tc.want, tc.tex)
		assert.NotContains(t, out, "<none></none>", tc.tex)
	}
}

// An environment whose body is a single node keeps that node in one cell.
func TestSingleNodeEnvironment(t *testing.T) {
	for _, tc := range []struct{ tex, want string }{
		{`\begin{matrix} \frac{a}{b} \end{matrix}`, `<mtd><mfrac><mi>a</mi><mi>b</mi></mfrac></mtd>`},
		{`\begin{pmatrix} \frac{1}{2} \end{pmatrix}`, `<mtd columnalign="center"><mfrac><mn>1</mn><mn>2</mn></mfrac></mtd>`},
		{`\begin{aligned} \sqrt{2} \end{aligned}`, `<msqrt><mn>2</mn></msqrt>`},
		{`\begin{array}{|c|} x \end{array}`, `<mtd columnalign="center"><mi>x</mi></mtd>`},
		{`\begin{array}{c} \sqrt{x} \end{array}`, `<mtd columnalign="center"><msqrt><mi>x</mi></msqrt></mtd>`},
	} {
		assert.Contains(t, mathBody(t, tc.tex), tc.want, tc.tex)
	}
}

// \over, \atop and \choose take everything before them in the group (or
// cell) as numerator and everything after as denominator.
func TestInfixScope(t *testing.T) {
	for _, tc := range []struct{ tex, want string }{
		{`x^2 \over 2`, `<mfrac><msup><mi>x</mi><mn>2</mn></msup><mn>2</mn></mfrac>`},
		{`a + b \over c`, `<mfrac><mrow><mi>a</mi><mo>+</mo><mi>b</mi></mrow><mi>c</mi></mfrac>`},
		{`{a \over b + c}`, `<mfrac><mi>a</mi><mrow><mi>b</mi><mo>+</mo><mi>c</mi></mrow></mfrac>`},
		{`1 \over \int\limits_0^1 f`, `<mfrac><mn>1</mn><mrow><mrow><munderover>`},
		{`n \choose k`, `<mfrac linethickness="0"><mi>n</mi><mi>k</mi></mfrac>`},
		{`\begin{matrix} a & b \over c \end{matrix}`, `<mtd><mfrac><mi>b</mi><mi>c</mi></mfrac></mtd>`},
	} {
		assert.Contains(t, mathBody(t, tc.tex), tc.want, tc.tex)
	}
	// TeX rejects two in one group; the error does not nest fractions.
	out := mathBody(t, strings.Repeat(`a \over `, 500)+"b")
	assert.Contains(t, out, `<merror title="ambiguous fraction: add braces"><mfrac><mi>a</mi>`)
	assert.Equal(t, 1, strings.Count(out, "<mfrac>"))
}

// A script at the start of a cell or row has no base: it does not attach to
// the separator before it and merge two cells or rows.
func TestScriptAtCellStart(t *testing.T) {
	for _, tc := range []struct {
		tex      string
		mtr, mtd int
	}{
		{`\begin{matrix} a & ^2 b \\ c & d \end{matrix}`, 2, 4},
		{`\begin{matrix} a \\ _1 b \end{matrix}`, 2, 2},
	} {
		out := mathBody(t, tc.tex)
		assert.Equal(t, tc.mtr, strings.Count(out, "<mtr"), tc.tex)
		assert.Equal(t, tc.mtd, strings.Count(out, "<mtd"), tc.tex)
	}
}

// \\ outside an environment breaks the line, it is not a literal backslash.
func TestLineBreakOutsideEnvironment(t *testing.T) {
	out := mathBody(t, `a = 1 \\ b = 2`)
	assert.NotContains(t, out, `<mo>\</mo>`)
	assert.Equal(t, 2, strings.Count(out, "<mtr"))
	// Inside a group it does not split the formula into rows.
	out = mathBody(t, `\frac{a \\ b}{c}`)
	assert.NotContains(t, out, `<mo>\</mo>`)
	assert.NotContains(t, out, "<mtr")
}

// A % comment does not change what follows it.
func TestCommentKeepsTheWindow(t *testing.T) {
	for _, p := range [][2]string{
		{"\\begin{pmatrix}%c\na & b\\end{pmatrix}.", "\\begin{pmatrix}a & b\\end{pmatrix}."},
		{"\\left( \\begin{matrix}%c\na & b\\end{matrix}\\right)", "\\left( \\begin{matrix}a & b\\end{matrix}\\right)"},
		{"\\begin{matrix}%c\na\\end{matrix}", "\\begin{matrix}a\\end{matrix}"},
		{"\\left(%c\nx\\right)", "\\left(x\\right)"},
		// A comment after a style switch in a cell.
		{"\\begin{matrix}\\bf %c\na & b\\end{matrix}", "\\begin{matrix}\\bf a & b\\end{matrix}"},
		{"\\begin{matrix}a\\color{red}%c\n & b\\end{matrix}", "\\begin{matrix}a\\color{red} & b\\end{matrix}"},
		// A comment whose text is a command name is still a comment.
		{"x %end", "x"},
		{"x %begin\n+y", "x+y"},
	} {
		assert.Equal(t, mathBody(t, p[1]), mathBody(t, p[0]), p[0])
	}
}

// TeX skips spaces between \left, \right, \middle or \big and the delimiter.
func TestDelimiterAfterSpace(t *testing.T) {
	assert.Equal(t, mathBody(t, `\left(x\right)`), mathBody(t, `\left ( x \right )`))
	assert.Equal(t, mathBody(t, `\big(x\big)`), mathBody(t, `\big ( x \big )`))
	assert.Equal(t, mathBody(t, `\left.x\middle|y\right.`), mathBody(t, `\left . x \middle | y \right .`))
}

// \left. and \right. draw nothing.
func TestNullDelimiter(t *testing.T) {
	for _, tex := range []string{`\left. x \right|`, `\left| x \right.`, `\left( x \middle. y \right)`} {
		out := mathBody(t, tex)
		assert.NotContains(t, out, `></mo>`, tex)
	}
}

func TestMathop(t *testing.T) {
	out := mathBody(t, `\mathop{\mathrm{argmax}}_x f`)
	assert.Contains(t, out, `<munder><mo`)
	assert.Contains(t, out, `>argmax</mo><mi>x</mi></munder>`)
	assert.NotContains(t, out, `mathrm`)
	out = mathBody(t, `\mathop{\sum}_i`)
	assert.Contains(t, out, `>∑</mo><mi>i</mi></munder>`)
}

// \bmod is an operator between its operands: it takes no argument.
func TestBmod(t *testing.T) {
	out := mathBody(t, `a \bmod \frac{p}{q}`)
	assert.NotContains(t, out, "merror")
	assert.Contains(t, out, `>mod</mo><mfrac><mi>p</mi><mi>q</mi></mfrac>`)
	// \pmod still takes one.
	assert.Contains(t, mathBody(t, `a \pmod{p}`), `>mod</mo><mi>p</mi><mo>)</mo>`)
}

// The \raisebox bound holds in the units of the math around it, whatever
// the size switches in between.
func TestRaiseboxUnderSizeSwitch(t *testing.T) {
	assert.NotContains(t, mathBody(t, `\Huge\raisebox{2em}{x}`), "voffset")
	// The inner shift would take the sum to 2.49em.
	assert.Equal(t, 1, strings.Count(mathBody(t, `\Huge\raisebox{0.5em}{\raisebox{0.5em}{x}}`), "voffset"))
	assert.Contains(t, mathBody(t, `\Huge\raisebox{0.5em}{x}`), `voffset="0.5em"`)
	// A length in pt does not grow with the font.
	assert.Contains(t, mathBody(t, `\LARGE\raisebox{20pt}{x}`), `voffset="20pt"`)
	assert.Contains(t, mathBody(t, `\tiny\raisebox{3em}{x}`), `voffset="3em"`)
}

// A run of switches in one group nests no deeper than the group.
func TestSwitchRunDepth(t *testing.T) {
	for _, unit := range []string{`\color{red}x+`, `\bf x+`, `\Large x+`} {
		tex := strings.Repeat(unit, 200) + "c"
		_, err := convertErrWithin(t, tex)
		assert.NoError(t, err, unit)
	}
}

func TestStarredMatrixColumnOption(t *testing.T) {
	out := mathBody(t, `\begin{pmatrix*}[r] 10 & 2 \\ 1 & 20 \end{pmatrix*}`)
	assert.NotContains(t, out, "[")
	assert.Contains(t, out, `<mtd columnalign="right"><mn>10</mn></mtd>`)
	// The brace form still works, and array skips its [position].
	assert.Contains(t, mathBody(t, `\begin{pmatrix*}{r} 10 & 2 \end{pmatrix*}`), `<mtd columnalign="right"><mn>10</mn></mtd>`)
	out = mathBody(t, `\begin{array}[t]{rl} 10 & 2 \end{array}`)
	assert.NotContains(t, out, "<mi>t</mi>")
	assert.Contains(t, out, `<mtd columnalign="right"><mn>10</mn></mtd>`)
}

func TestColorModel(t *testing.T) {
	out := mathBody(t, `\color[RGB]{255,0,0}{x}`)
	assert.NotContains(t, out, "merror")
	assert.NotContains(t, out, "<mn>255</mn>")
	assert.Contains(t, out, "<mi>x</mi>")
}

func TestNewcommandArgCount(t *testing.T) {
	for _, tex := range []string{`\newcommand{\f}[1]{#1^2} x`, `\newcommand{\f}[2][a]{#1+#2} x`} {
		out := mathBody(t, tex)
		assert.NotContains(t, out, "]", tex)
		assert.NotContains(t, out, "<mn>1</mn>", tex)
		assert.Contains(t, out, "<mi>x</mi>", tex)
	}
}

func TestSidesetRepeatedScripts(t *testing.T) {
	out := mathBody(t, `\sideset{_a_b}{}\sum`)
	assert.Contains(t, out, `<mi>a</mi><none></none><mi>b</mi><none></none>`)
	out = mathBody(t, `\sideset{^a^b}{}\sum`)
	assert.Contains(t, out, `<none></none><mi>a</mi><none></none><mi>b</mi>`)
}

func TestDoubleColumnRule(t *testing.T) {
	_, lines := parseAlignmentString("c||cc")
	assert.Equal(t, []string{"solid", "none"}, lines)
}

// Only a \\[len] sets the spacing of its row.
func TestSubstackRowspacingStaysInside(t *testing.T) {
	out := mathBody(t, `\begin{matrix} \substack{a\\b} & y \\ c & d \end{matrix}`)
	assert.Equal(t, 1, strings.Count(out, `rowspacing=`))
	assert.Contains(t, mathBody(t, `\begin{matrix} a \\[2em] b \end{matrix}`), `rowspacing="2em`)
}

func TestNewMMLNodeExtraArgs(t *testing.T) {
	n := NewMMLNode("mi", "x", "ignored")
	assert.Equal(t, "mi", n.Tag)
	assert.Equal(t, "x", n.Text)
}

func TestDepthErrorIsReported(t *testing.T) {
	_, err := NewMathMLConverter().ConvertInline(strings.Repeat("{", 100) + "x" + strings.Repeat("}", 100))
	assert.ErrorIs(t, err, errMaxDepth)
}

func TestBraceErrorHasNoMarkup(t *testing.T) {
	_, err := NewMathMLConverter().ConvertInline(`{<b>x`)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "<pre>")
}
