// This file serves as the entry point to load the Gno Goldmark extension.
// Goldmark extensions are designed as follows:
//
//  <Markdown in []byte, parser.Context>
//                 |
//                 V
//  +-------- parser.Parser ---------------------------
//  | 1. Parse block elements into AST
//  |   1. If a parsed block is a paragraph, apply
//  |      ast.ParagraphTransformer
//  | 2. Traverse AST and parse blocks.
//  |   1. Process delimiters (emphasis) at the end of
//  |      block parsing
//  | 3. Apply parser.ASTTransformers to AST
//                 |
//                 V
//            <ast.Node>
//                 |
//                 V
//  +------- renderer.Renderer ------------------------
//  | 1. Traverse AST and apply renderer.NodeRenderer
//  |    corresponding to the node type
//
//                 |
//                 V
//              <Output>
//
// More information can be found on the Goldmark repository page:
// https://github.com/yuin/goldmark#goldmark-internalfor-extension-developers

package markdown

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

var _ goldmark.Extender = (*gnoExtension)(nil)

// NewGnoParserContext creates a new parser context with GnoURL
func NewGnoParserContext(url *weburl.GnoURL) parser.Context {
	ctx := parser.NewContext()
	ctx.Set(gUrlContextKey, *url)
	return ctx
}

var gUrlContextKey = parser.NewContextKey()

// getUrlFromContext retrieves the GnoURL from the parser context
func getUrlFromContext(ctx parser.Context) (url weburl.GnoURL, ok bool) {
	url, ok = ctx.Get(gUrlContextKey).(weburl.GnoURL)
	return
}

type gnoExtension struct{}

var GnoExtension = &gnoExtension{}

// Extend adds the Gno extension to the provided Goldmark markdown processor.
func (e *gnoExtension) Extend(m goldmark.Markdown) {
	// Add column extension
	Columns.Extend(m)

	// Add link extension with context
	Links.Extend(m)
}
