package markdown

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// A run of openers must stop at the shared nesting cap: goldmark keeps
// per-line state for every open block, so unbounded nesting costs time and
// memory quadratic in the input.
func TestDetailsNestBounded(t *testing.T) {
	u, err := weburl.Parse("https://gno.land/r/test")
	require.NoError(t, err)
	m := goldmark.New()
	NewGnoExtension().Extend(m)

	src := []byte(strings.Repeat(":::details a\n", 4000))
	doc := m.Parser().Parse(text.NewReader(src),
		parser.WithContext(NewGnoParserContext(GnoContext{GnoURL: u})))

	count := 0
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == KindDetailsBlock {
			count++
		}
		return ast.WalkContinue, nil
	})
	require.Equal(t, MaxGnoNestDepth, count)
}
