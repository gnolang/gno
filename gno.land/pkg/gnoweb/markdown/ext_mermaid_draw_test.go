package markdown

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func TestMermaidDrawn(t *testing.T) {
	html, nodes := renderMermaidDoc(t, seqFence, true)
	assert.Equal(t, 1, nodes)
	assert.Contains(t, html, `<div class="gno-mermaid"><div data-copy-exclude><svg`)
	// The source follows, visually hidden, for assistive technology and Copy.
	assert.Contains(t, html, `</svg></div><pre class="u-sr-only">sequenceDiagram
A-&gt;&gt;B: hi
</pre></div>`)
	assert.NotContains(t, html, "<pre><code")
}

func TestMermaidPerDocument(t *testing.T) {
	// At most maxMermaidTries fences are tried; the ones past it show as
	// code, as do the ones that fail.
	html, nodes := renderMermaidDoc(t, strings.Repeat(seqFence, maxMermaidTries+2), true)
	assert.Equal(t, maxMermaidTries+2, nodes)
	assert.Equal(t, maxMermaidTries, strings.Count(html, "<svg"))
	assert.Equal(t, 2, strings.Count(html, "<pre><code"))

	broken := strings.Repeat("```mermaid\nflowchart LR\n```\n\n", 3)
	html, _ = renderMermaidDoc(t, broken+seqFence, true)
	assert.Equal(t, 1, strings.Count(html, "<svg"))
	assert.Equal(t, 3, strings.Count(html, "<pre><code"))
}

func TestMermaidUnrenderedFootnote(t *testing.T) {
	// A fence in a footnote never referenced is not rendered, so it takes
	// none of the document's tries.
	m := goldmark.New(goldmark.WithExtensions(extension.Footnote))
	NewGnoExtension().Extend(m)
	unused := ""
	for i := range maxMermaidTries {
		unused += fmt.Sprintf("[^u%d]:\n    ```mermaid\n    sequenceDiagram\n    A->>B: hi\n    ```\n\n", i)
	}
	src := []byte(unused + seqFence)
	doc := m.Parser().Parse(text.NewReader(src), parser.WithContext(NewGnoParserContext(GnoContext{Mermaid: true})))
	var out bytes.Buffer
	require.NoError(t, m.Renderer().Render(&out, src, doc))
	assert.Equal(t, 1, strings.Count(out.String(), "<svg"))
}
