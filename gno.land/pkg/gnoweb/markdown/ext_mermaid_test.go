package markdown

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// renderMermaidDoc renders src with the Gno extension, diagrams enabled or
// not, and returns the HTML and the number of mermaid nodes.
func renderMermaidDoc(t *testing.T, src string, enabled bool) (string, int) {
	t.Helper()
	m := goldmark.New()
	NewGnoExtension().Extend(m)
	ctx := NewGnoParserContext(GnoContext{Mermaid: enabled})
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

func TestMermaidDisabled(t *testing.T) {
	html, nodes := renderMermaidDoc(t, seqFence, false)
	assert.Zero(t, nodes, "a disabled context leaves the fence alone")
	assert.Contains(t, html, "<pre><code")
	assert.NotContains(t, html, "<svg")
}

func TestMermaidFences(t *testing.T) {
	// Like ```math: the info string is exactly "mermaid", and a blank fence
	// stays a code block.
	_, nodes := renderMermaidDoc(t, "```mermaid title\nsequenceDiagram\nA->>B: hi\n```\n\n```mermaid\n\n```\n\n"+seqFence, true)
	assert.Equal(t, 1, nodes)
}

func TestMermaidFallback(t *testing.T) {
	// A source the package does not draw shows as the code block it is.
	html, nodes := renderMermaidDoc(t, "```mermaid\nunknownDiagram\n  <b>x</b>\n```\n", true)
	require.Equal(t, 1, nodes)
	assert.Contains(t, html, "<pre><code")
	assert.Contains(t, html, "&lt;b&gt;x&lt;/b&gt;")
	assert.NotContains(t, html, "gno-mermaid")
}

func TestMermaidNotInForeign(t *testing.T) {
	// The <gno-foreign> sandbox does not load the extension: a fence there
	// stays code, even with diagrams enabled.
	html, _ := renderMermaidDoc(t, "<gno-foreign>\n\n"+seqFence+"</gno-foreign>\n", true)
	assert.Contains(t, html, "language-mermaid")
	assert.NotContains(t, html, "<svg")
}
