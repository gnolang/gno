package markdown

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// ImageValidatorFunc validates image URLs. It should return `true` for any valid image URL.
type ImageValidatorFunc func(uri string) (ok bool)

// imgValidatorTransformer implements ASTTransformer
type imgValidatorTransformer struct {
	valFunc ImageValidatorFunc
}

// Transform iterate on `ast.Image` nodes and validate images URLs.
func (t *imgValidatorTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	if t.valFunc == nil {
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

		if !t.valFunc(string(img.Destination)) {
			img.Destination = []byte{} // Erase destination
		}

		return ast.WalkContinue, nil
	})
}

type imgValidatorExtension struct{}

// ExtImageValidator is a Goldmark extension that pre validation on image URLs.
var ExtImageValidator = &imgValidatorExtension{}

// Extend adds the ExtImageValidator to the provided Goldmark markdown processor
func (l *imgValidatorExtension) Extend(m goldmark.Markdown, valFunc ImageValidatorFunc) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&imgValidatorTransformer{valFunc}, 500),
	))
}
