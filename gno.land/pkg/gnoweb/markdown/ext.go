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
	"github.com/yuin/goldmark"
)

var _ goldmark.Extender = (*gnoExtension)(nil)

type gnoExtension struct {
	domain string
	path   string
}

// NewGnoExtension creates a new GnoExtension instance
func NewGnoExtension(domain, path string) goldmark.Extender {
	return &gnoExtension{
		domain: domain,
		path:   path,
	}
}

// Extend adds the Gno extension to the provided Goldmark markdown processor.
func (e *gnoExtension) Extend(m goldmark.Markdown) {
	Columns.Extend(m)
	Links.Extend(m, e.domain, e.path)
}
