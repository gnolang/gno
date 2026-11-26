package markdown

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMermaidDrawn(t *testing.T) {
	html, nodes := renderMermaidDoc(t, seqFence, true)
	assert.Equal(t, 1, nodes)
	assert.Contains(t, html, `<div class="gno-mermaid"><svg`)
	assert.NotContains(t, html, "<pre")
}

func TestMermaidPerDocument(t *testing.T) {
	// At most maxMermaidTries fences are tried; the ones past it show as
	// code, as do the ones that fail.
	html, nodes := renderMermaidDoc(t, strings.Repeat(seqFence, maxMermaidTries+2), true)
	assert.Equal(t, maxMermaidTries, nodes)
	assert.Equal(t, maxMermaidTries, strings.Count(html, "<svg"))
	assert.Equal(t, 2, strings.Count(html, "<pre><code"))

	broken := strings.Repeat("```mermaid\nflowchart LR\n```\n\n", 3)
	html, _ = renderMermaidDoc(t, broken+seqFence, true)
	assert.Equal(t, 1, strings.Count(html, "<svg"))
	assert.Equal(t, 3, strings.Count(html, "<pre><code"))
}
