package markdown

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
)

func renderMathMarkdown(t *testing.T, src string) string {
	t.Helper()
	var buf bytes.Buffer
	gm := goldmark.New(goldmark.WithExtensions(NewGnoExtension()))
	require.NoError(t, gm.Convert([]byte(src), &buf))
	return buf.String()
}

func TestMathEscapesUserContent(t *testing.T) {
	cases := map[string]string{
		"text breaks out of math":     `$$\text{</math><script>alert('XSS')</script>}$$`,
		"class attribute injection":   `$\class{x" onclick="alert(1)}{y}$`,
		"color attribute injection":   `$\color{red" onmouseover="alert(1)}{y}$`,
		"fallback breaks out of span": `$\begin{matrix}</span><script>alert(1)</script>$`,
		"fallback breaks out of div":  "$$\n\\begin{matrix}</div><script>alert(1)</script>\n$$",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			out := renderMathMarkdown(t, src)
			assert.NotContains(t, out, "<script")
			assert.NotContains(t, out, `" onclick=`)
			assert.NotContains(t, out, `" onmouseover=`)
		})
	}
}
