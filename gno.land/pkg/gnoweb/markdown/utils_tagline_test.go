package markdown

import (
	"testing"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// TestGnoTagLineParserOpensOnPrefix covers the shared gnoTagLineParser,
// which this branch carries for the inline gno-* tags of sibling branches:
// it opens a paragraph on the tag name alone, and on nothing else.
func TestGnoTagLineParserOpensOnPrefix(t *testing.T) {
	p := newGnoTagLineParser([]byte("<gno-x"))
	cases := map[string]bool{
		"<gno-x />\n":               true,
		"  <GNO-X a=\"1\">\n":       true,
		"<gno-x a=\"unterminated\n": true, // tag name alone is enough
		"<gno-xy />\n":              false,
		"text\n":                    false,
	}
	for line, want := range cases {
		node, _ := p.Open(ast.NewDocument(), text.NewReader([]byte(line)), parser.NewContext())
		if got := node != nil; got != want {
			t.Errorf("Open(%q) opened = %v, want %v", line, got, want)
		}
	}
}
