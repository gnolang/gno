package markdown

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// parseDetails parses src with the gno extensions and returns the document,
// its HTML and the gno-* nesting depth left on the parser context.
func parseDetails(t *testing.T, src string) (ast.Node, string, int) {
	t.Helper()
	m := newGnoMarkdown()
	pc := NewGnoParserContext(GnoContext{})
	doc := m.Parser().Parse(text.NewReader([]byte(src)), parser.WithContext(pc))
	var buf bytes.Buffer
	require.NoError(t, m.Renderer().Render(&buf, []byte(src), doc))
	return doc, buf.String(), Get(pc)
}

// A run of openers must stop at the shared nesting cap: goldmark keeps
// per-line state for every open block, so unbounded nesting costs time and
// memory quadratic in the input.
func TestDetailsNestBounded(t *testing.T) {
	doc, _, _ := parseDetails(t, strings.Repeat(":::details a\n", 4000))
	count := 0
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == KindDetailsBlock {
			count++
		}
		return ast.WalkContinue, nil
	})
	require.Equal(t, MaxGnoNestDepth, count)
}

// Every Push in Open is matched by one Pop, whatever shape closes the block.
func TestDetailsDepthBalanced(t *testing.T) {
	cases := map[string]string{
		"closed":        ":::details a\nb\n:::\n",
		"unclosed":      ":::details a\nb\n",
		"over_cap":      strings.Repeat(":::details a\n", 10),
		"in_quote_lazy": "> :::details a\n> b\n\nafter\n",
		"in_list_ends":  "- :::details a\n  b\n\npara\n",
		"sequential":    strings.Repeat(":::details a\nb\n:::\n", 6),
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			_, _, depth := parseDetails(t, src)
			require.Zero(t, depth)
		})
	}
}

// A container that consumes part of a tab leaves padding on the line, which
// the summary offsets must skip.
func TestDetailsTabPadding(t *testing.T) {
	cases := map[string]string{
		"quote_tab": ">\t:::details Summary text\n>\tbody\n>\t:::\n",
		"alert_tab": "> [!NOTE]\n>\t:::details Summary text\n>\tbody\n>\t:::\n",
		"list_tab":  "- item\n\t:::details Summary text\n\tbody\n\t:::\n",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			_, out, _ := parseDetails(t, src)
			require.Contains(t, out, "</svg>Summary text</summary>")
		})
	}
}
