package markdown

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
)

// TestLinearIDsMatchesDefault feeds the same sequence of Generate and Put
// calls to goldmark's default generator and to linearIDs: the ids must be
// identical, including suffix collisions with explicit ids and with headings
// whose text already ends in "-N".
func TestLinearIDsMatchesDefault(t *testing.T) {
	type call struct {
		put   string
		value string
		kind  ast.NodeKind
	}
	h, p := ast.KindHeading, ast.KindParagraph
	calls := []call{
		{value: "a", kind: h}, {value: "a", kind: h}, {value: "a-1", kind: h},
		{value: "a", kind: h}, {put: "a-4"}, {value: "a", kind: h},
		{value: "A", kind: h}, {value: " a ", kind: h}, {value: "", kind: h},
		{value: "", kind: h}, {value: "", kind: p}, {value: "é", kind: h},
		{value: "Hello World_x", kind: h}, {value: "hello-world-x", kind: h},
		{put: "b"}, {value: "b", kind: h}, {value: "b", kind: h},
		{value: "<gno-icon name=\"x\" />", kind: h}, {value: "gno-icon-namex-", kind: h},
	}
	def := parser.NewContext().IDs()
	lin := newLinearIDs()
	for i, c := range calls {
		if c.put != "" {
			def.Put([]byte(c.put))
			lin.Put([]byte(c.put))
			continue
		}
		want := string(def.Generate([]byte(c.value), c.kind))
		got := string(lin.Generate([]byte(c.value), c.kind))
		require.Equal(t, want, got, "call %d (%q)", i, c.value)
	}
}

// TestDuplicateHeadingIDsLinear guards the render time of many identical
// headings: with goldmark's default generator 20k "## a" lines took tens of
// seconds.
func TestDuplicateHeadingIDsLinear(t *testing.T) {
	md := goldmark.New(
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithExtensions(NewGnoExtension()),
	)
	src := []byte(strings.Repeat("## a\n", 20000))
	var out bytes.Buffer
	start := time.Now()
	require.NoError(t, md.Convert(src, &out, parser.WithContext(NewGnoParserContext(GnoContext{}))))
	require.Less(t, time.Since(start), 2*time.Second)
	require.Contains(t, out.String(), `<h2 id="a-19999">a</h2>`)
}
