package markdown

import (
	"bytes"
	"testing"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
)

func newTestDocMarkdown(t *testing.T) goldmark.Markdown {
	t.Helper()
	formatter := html.New(html.WithClasses(true), html.ClassPrefix("chroma-"))
	style := styles.Get("friendly")
	return goldmark.New(goldmark.WithExtensions(ExtCodeExpand(formatter, style)))
}

func TestExtCodeExpand_FencedGoBlock(t *testing.T) {
	t.Parallel()
	md := newTestDocMarkdown(t)
	input := "```go\nfmt.Println(\"hi\")\n```\n"
	var buf bytes.Buffer
	require.NoError(t, md.Convert([]byte(input), &buf))
	out := buf.String()
	require.Contains(t, out, `<details class="doc-example">`)
	require.Contains(t, out, `<summary>Example</summary>`)
	require.Contains(t, out, `class="chroma-`) // chroma highlight classes present
	require.Contains(t, out, `</details>`)
}

func TestExtCodeExpand_FencedBlockNoLangFallsBackToPlainText(t *testing.T) {
	t.Parallel()
	md := newTestDocMarkdown(t)
	input := "```\nfmt.Println(\"hi\")\n```\n"
	var buf bytes.Buffer
	require.NoError(t, md.Convert([]byte(input), &buf))
	out := buf.String()
	require.Contains(t, out, `<details class="doc-example">`)
	require.Contains(t, out, `class="chroma-`)
}

func TestExtCodeExpand_IndentedBlockRendersSameAsFenced(t *testing.T) {
	t.Parallel()
	md := newTestDocMarkdown(t)
	// Four-space-indented block.
	input := "Intro.\n\n    fmt.Println(\"hi\")\n"
	var buf bytes.Buffer
	require.NoError(t, md.Convert([]byte(input), &buf))
	out := buf.String()
	require.Contains(t, out, `<details class="doc-example">`,
		"indented code block must render as collapsible details, same as fenced")
}

func TestExtCodeExpand_UnknownLangFallsBackToPlainText(t *testing.T) {
	t.Parallel()
	md := newTestDocMarkdown(t)
	input := "```rust\nfn main() {}\n```\n"
	var buf bytes.Buffer
	require.NoError(t, md.Convert([]byte(input), &buf))
	out := buf.String()
	require.Contains(t, out, `<details class="doc-example">`)
	// Unknown language falls back to the plain-text lexer, which still goes
	// through chromahtml.Formatter — output carries chroma classes.
	require.Contains(t, out, `class="chroma-`)
}

func TestExtCodeExpand_EscapesHTMLInsideCode(t *testing.T) {
	t.Parallel()
	md := newTestDocMarkdown(t)
	input := "```\n<script>alert(1)</script>\n```\n"
	var buf bytes.Buffer
	require.NoError(t, md.Convert([]byte(input), &buf))
	out := buf.String()
	require.NotContains(t, out, "<script>alert(1)</script>", "raw script tag must not survive")
	require.Contains(t, out, "&lt;script&gt;", "html must be escaped in code content")
}

func TestExtCodeExpand_TextAroundCodeBlocksUnchanged(t *testing.T) {
	t.Parallel()
	md := newTestDocMarkdown(t)
	input := "Hello **world**.\n\n```go\nx := 1\n```\n\nGoodbye.\n"
	var buf bytes.Buffer
	require.NoError(t, md.Convert([]byte(input), &buf))
	out := buf.String()
	require.Contains(t, out, "<strong>world</strong>")
	require.Contains(t, out, `<details class="doc-example">`)
	require.Contains(t, out, "Goodbye")
}
