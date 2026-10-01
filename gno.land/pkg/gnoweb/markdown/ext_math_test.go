package markdown

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

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

func TestMathDoesNotSwallowText(t *testing.T) {
	cases := map[string]string{
		`\alpha is greek`:                    `\alpha is greek`,
		`$100 is the price`:                  `$100 is the price`,
		`\_underscore`:                       `_underscore`,
		`$5 and $10`:                         `$5 and $10`,
		`costs \$5 and \$6`:                  `costs $5 and $6`,
		"$$\nnever closed\n\nnext paragraph": "next paragraph",
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			out := renderMathMarkdown(t, src)
			assert.Contains(t, out, want)
			assert.NotContains(t, out, "<math")
		})
	}
}

func TestMathStillRenders(t *testing.T) {
	for _, src := range []string{
		`$E=mc^2$`,
		`price $x^2$ here`,
		`$$\int_0^1 x^2 dx$$`,
		"$$\n\\int_0^1 x^2 dx\n$$",
		`\\(A = \\pi r^2\\)`,
		`a \\$x$ b`,
		`$a\\$`,
	} {
		assert.Contains(t, renderMathMarkdown(t, src), "<math", src)
	}
}

// Run with -race: the renderer used to share one converter across renders.
func TestMathConcurrentRender(t *testing.T) {
	gm := goldmark.New(goldmark.WithExtensions(NewGnoExtension()))
	in := [][]byte{[]byte(`$E=mc^2$`), []byte(`$$\int_0^1 x^2 dx$$`)}
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var b bytes.Buffer
			assert.NoError(t, gm.Convert(in[i%2], &b))
			assert.Contains(t, b.String(), "<math")
		}()
	}
	wg.Wait()
}

func TestMathOutputIsBounded(t *testing.T) {
	t.Run("deep nesting", func(t *testing.T) {
		for _, depth := range []int{50, 1000} {
			src := "$" + strings.Repeat(`\sqrt{`, depth) + "x" + strings.Repeat("}", depth) + "$"
			out := renderMathMarkdown(t, src)
			assert.Less(t, len(out), 64*len(src)+4096, "depth %d", depth)
		}
	})
	t.Run("many unclosed openers", func(t *testing.T) {
		// Each unclosed opener looks ahead for a closing delimiter; the
		// lookahead is cached so this stays linear.
		start := time.Now()
		out := renderMathMarkdown(t, strings.Repeat("\\\\[\n", 1<<16))
		assert.NotContains(t, out, "<math")
		assert.Less(t, time.Since(start), 10*time.Second)
	})
	t.Run("too long", func(t *testing.T) {
		src := "$" + strings.Repeat("x+", MaxMathInputLen) + "x$"
		out := renderMathMarkdown(t, src)
		assert.NotContains(t, out, "<math")
		assert.Contains(t, out, `<span class="math-inline">`)
	})
}
