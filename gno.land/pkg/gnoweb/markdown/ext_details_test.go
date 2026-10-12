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

// A run of openers stops at the shared nesting cap; the rest are refused.
func TestDetailsNestBounded(t *testing.T) {
	doc, _, _ := parseDetails(t, strings.Repeat("<gno-details summary=\"a\">\n", 4000))
	count := 0
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if d, ok := n.(*GnoDetailsNode); ok && entering && d.tag == detailsTagOpen {
			count++
		}
		return ast.WalkContinue, nil
	})
	require.Equal(t, MaxGnoNestDepth, count)
}

// Every Push in Open is matched by one Pop, whatever shape ends the block.
func TestDetailsDepthBalanced(t *testing.T) {
	cases := map[string]string{
		"closed":          "<gno-details>\nb\n</gno-details>\n",
		"unclosed":        "<gno-details>\nb\n",
		"over_cap":        strings.Repeat("<gno-details>\n", 10),
		"in_quote_ends":   "> <gno-details>\n> b\n\nafter\n",
		"in_list_ends":    "- <gno-details>\n  b\n\npara\n",
		"inner_unclosed":  "<gno-details>\n<gno-details>\nb\n</gno-details>\n",
		"close_elsewhere": "<gno-details>\n- item\n  </gno-details>\n",
		"sequential":      strings.Repeat("<gno-details>\nb\n</gno-details>\n", 6),
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			_, _, depth := parseDetails(t, src)
			require.Zero(t, depth)
		})
	}
}

// Closed blocks give their depth back while parsing, so a gno-* block
// after four of them still opens.
func TestDetailsSequentialKeepsCap(t *testing.T) {
	src := strings.Repeat("<gno-details>\nb\n</gno-details>\n", 5) + "> [!NOTE]\n> x\n"
	_, out, _ := parseDetails(t, src)
	require.Contains(t, out, "gno-alert")
}
