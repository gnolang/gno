package markdown

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown/mermaid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// renderMermaidDoc renders src with the Gno extension and returns the HTML
// and the number of mermaid nodes.
func renderMermaidDoc(t *testing.T, src string) (string, int) {
	t.Helper()
	m := goldmark.New()
	NewGnoExtension().Extend(m)
	ctx := NewGnoParserContext(GnoContext{})
	doc := m.Parser().Parse(text.NewReader([]byte(src)), parser.WithContext(ctx))
	nodes := 0
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == KindMermaid {
			nodes++
		}
		return ast.WalkContinue, nil
	})
	var out bytes.Buffer
	require.NoError(t, m.Renderer().Render(&out, []byte(src), doc))
	return out.String(), nodes
}

const seqFence = "```mermaid\nsequenceDiagram\nA->>B: hi\n```\n\n"

func TestMermaidFences(t *testing.T) {
	// Like ```math: the info string is exactly "mermaid", and a blank fence
	// stays a code block.
	_, nodes := renderMermaidDoc(t, "```mermaid title\nsequenceDiagram\nA->>B: hi\n```\n\n```mermaid\n\n```\n\n"+seqFence)
	assert.Equal(t, 1, nodes)
}

func TestMermaidTag(t *testing.T) {
	const pie = "pie\n\"a\" : 1\n"
	// The golden file ext_mermaid/tag.md covers each width and an unknown one.
	for name, c := range map[string]struct{ src, want string }{
		"tag case":      {"<GNO-MERMAID width=full>\n" + pie + "</gno-mermaid >\n", `<div class="gno-mermaid gno-mermaid--full"><div class="gno-mermaid__scale" data-copy-exclude><svg`},
		"in a list":     {"- x\n\n  <gno-mermaid>\n  pie\n  \"a\" : 1\n  </gno-mermaid>\n", `<div class="gno-mermaid"><div class="gno-mermaid__scale" data-copy-exclude><svg`},
		"after text":    {"Votes:\n<gno-mermaid>\n" + pie + "</gno-mermaid>\n", "<p>Votes:</p>\n<div class=\"gno-mermaid\"><div class=\"gno-mermaid__scale\" data-copy-exclude><svg"},
		"bad attribute": {"<gno-mermaid width=fit style=x>\n" + pie + "</gno-mermaid>\n", "<pre><code>pie"},
		"unclosed":      {"<gno-mermaid>\n" + pie, "<pre><code>pie"},
		"list ends":     {"- <gno-mermaid>\n  " + pie + "\nafter\n", "<pre><code>pie"},
		"not drawable":  {"<gno-mermaid>\nflowchart LR\n<b>x</b>\n</gno-mermaid>\n", "&lt;b&gt;x&lt;/b&gt;"},
	} {
		t.Run(name, func(t *testing.T) {
			html, nodes := renderMermaidDoc(t, c.src)
			assert.Equal(t, 1, nodes)
			assert.Contains(t, html, c.want)
			assert.Equal(t, strings.Contains(c.want, "<svg"), strings.Contains(html, "<svg"), html)
			assert.NotContains(t, html, "gno-mermaid>", "the tags never reach the page")
		})
	}

}

func TestMermaidStrayTag(t *testing.T) {
	// A gno-mermaid tag line that opens no block renders as a comment,
	// rather than turning into an HTML block that hides the lines after it.
	for name, c := range map[string]struct {
		src  string
		want []string
	}{
		"close after its blockquote": {
			"> <gno-mermaid>\n> pie\n> \"a\" : 1\n</gno-mermaid>\nVote ends **tomorrow**.\nQuorum: 60%.\n\nFooter",
			[]string{"<p>Vote ends <strong>tomorrow</strong>.\nQuorum: 60%.</p>", "<p>Footer</p>"},
		},
		"self-closing": {
			"<gno-mermaid width=\"fit\"/>\npie\n\"a\" : 1\n\n\"b\" : 2\n</gno-mermaid>",
			[]string{"<p>pie\n&quot;a&quot; : 1</p>", "<p>&quot;b&quot; : 2</p>"},
		},
		"second close": {
			"<gno-mermaid>\npie\n\"a\" : 1\n</gno-mermaid>\n</gno-mermaid>\nafter\n",
			[]string{"<p>after</p>"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			html, _ := renderMermaidDoc(t, c.src)
			for _, want := range c.want {
				assert.Contains(t, html, want)
			}
			assert.Contains(t, html, "<!-- unexpected/invalid mermaid tag omitted -->")
			assert.NotContains(t, html, "raw HTML omitted")
		})
	}
}

func TestMermaidNotInForeign(t *testing.T) {
	// The <gno-foreign> sandbox does not load the extension: a fence there
	// stays code.
	html, _ := renderMermaidDoc(t, "<gno-foreign>\n\n"+seqFence+"</gno-foreign>\n")
	assert.Contains(t, html, "language-mermaid")
	assert.NotContains(t, html, "<svg")

	// A <gno-mermaid> block there is an unknown tag, omitted like any other.
	html, _ = renderMermaidDoc(t, "<gno-foreign>\n\n<gno-mermaid>\npie\n\"a\" : 1\n</gno-mermaid>\n</gno-foreign>\n")
	assert.NotContains(t, html, "<svg")
	assert.NotContains(t, html, "gno-mermaid>")
}

func TestMermaidPerDocument(t *testing.T) {
	// At most maxMermaidTries fences are tried; the ones past it show as
	// code, as do the ones that fail.
	html, nodes := renderMermaidDoc(t, strings.Repeat(seqFence, maxMermaidTries+2))
	assert.Equal(t, maxMermaidTries+2, nodes)
	assert.Equal(t, maxMermaidTries, strings.Count(html, `class="gno-mermaid__svg"`))
	assert.Equal(t, 2, strings.Count(html, "<pre><code"))

	broken := strings.Repeat("```mermaid\nflowchart LR\n```\n\n", 3)
	html, _ = renderMermaidDoc(t, broken+seqFence)
	assert.Equal(t, 1, strings.Count(html, `class="gno-mermaid__svg"`))
	assert.Equal(t, 3, strings.Count(html, "<pre><code"))
}

func TestMermaidBytesSpent(t *testing.T) {
	// The first diagram the page's SVG budget cannot hold spends it: it and
	// the ones after it show as code, without being drawn.
	var big strings.Builder
	big.WriteString("pie\n")
	for i := range 800 {
		fmt.Fprintf(&big, "\"%d\" : 1\n", i)
	}
	svg, _, err := mermaid.Render(big.String())
	require.NoError(t, err)
	fits := maxMermaidBytes / len(svg)
	require.Less(t, fits+2, maxMermaidTries)

	fence := "```mermaid\n" + big.String() + "```\n\n"
	html, _ := renderMermaidDoc(t, strings.Repeat(fence, fits+1)+seqFence)
	assert.Equal(t, fits, strings.Count(html, `class="gno-mermaid__svg"`))
	assert.Equal(t, 2, strings.Count(html, "<pre><code"))
}

func TestMermaidUnrenderedFootnote(t *testing.T) {
	// A fence in a footnote never referenced is not rendered, so it takes
	// none of the document's tries.
	m := goldmark.New(goldmark.WithExtensions(extension.Footnote))
	NewGnoExtension().Extend(m)
	var unused strings.Builder
	for i := range maxMermaidTries {
		unused.WriteString(fmt.Sprintf("[^u%d]:\n    ```mermaid\n    sequenceDiagram\n    A->>B: hi\n    ```\n\n", i))
	}
	src := []byte(unused.String() + seqFence)
	doc := m.Parser().Parse(text.NewReader(src), parser.WithContext(NewGnoParserContext(GnoContext{})))
	var out bytes.Buffer
	require.NoError(t, m.Renderer().Render(&out, src, doc))
	assert.Equal(t, 1, strings.Count(out.String(), `class="gno-mermaid__svg"`))
}
