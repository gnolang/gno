package markdown

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// ImageFilterFunc filter the element if returning `true`
type ImageFilterFunc func(uri string) bool

// linkTransformer implements ASTTransformer
type imgFilterTransformer struct {
	filterFunc ImageFilterFunc
}

// Transform filterr `ast.Image` nodes if `filterFunc` return true.
func (t *imgFilterTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	if t.filterFunc == nil {
		return
	}

	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		img, ok := node.(*ast.Image)
		if !ok {
			return ast.WalkContinue, nil
		}

		if t.filterFunc(string(img.Destination)) {
			img.Destination = []byte{} // Erase destination
		}

		return ast.WalkContinue, nil
	})
}

// imgFilterExtension is a Goldmark extension that handles link rendering with special attributes
// for external, internal, and same-package links.
type imgFilterExtension struct{}

// Links instance for extending markdown with link functionality
var ExtImageFilter = &imgFilterExtension{}

// Extend adds the LinkExtension to the provided Goldmark markdown processor
func (l *imgFilterExtension) Extend(m goldmark.Markdown, filter ImageFilterFunc) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&imgFilterTransformer{filter}, 500),
	))
}
