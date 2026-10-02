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

func TestMMLNodeWriteKeepsEntities(t *testing.T) {
	for _, s := range []string{"&OverBrace;", "&lt;", "&#8289;", "&#x2061;"} {
		var b strings.Builder
		NewMMLNode("mo", s).Write(&b, -1)
		assert.Equal(t, "<mo>"+s+"</mo>", b.String())
	}
	var b strings.Builder
	NewMMLNode("mo", "&notanentity &").Write(&b, -1)
	assert.Equal(t, "<mo>&amp;notanentity &amp;</mo>", b.String())
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
