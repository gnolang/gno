package mathml

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMMLNodeWriteEscapes(t *testing.T) {
	n := NewMMLNode("mi", `</math><script>x & y</script>`)
	n.SetAttr("class", `x" onclick="alert(1)`)
	n.SetAttr(`bad" onclick="alert(1)`, "v")
	n.SetCssProp("color", `red"><script>`)
	var b strings.Builder
	n.Write(&b, -1)
	assert.Equal(t,
		`<mi class="x&#34; onclick=&#34;alert(1)" style="color:red&#34;&gt;&lt;script&gt;;">&lt;/math&gt;&lt;script&gt;x &amp; y&lt;/script&gt;</mi>`,
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
		NewMMLNode("mo", in).Write(&b, -1)
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

func TestMMLNodeWriteSortsCSS(t *testing.T) {
	n := NewMMLNode("mo", "lim").SetCssProp("padding", "0").SetCssProp("border-bottom", "1px").SetCssProp("color", "red")
	for range 20 {
		var b strings.Builder
		n.Write(&b, -1)
		assert.Equal(t, `<mo style="border-bottom:1px;color:red;padding:0;">lim</mo>`, b.String())
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
