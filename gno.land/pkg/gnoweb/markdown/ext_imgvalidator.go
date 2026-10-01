package markdown

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// ImageValidatorFunc validates image URLs. It should return `true` for any valid image URL.
type ImageValidatorFunc func(uri string) (ok bool)

// AllowSvgDataImage is gnoweb's image policy: every data: URI is rejected
// except image/svg+xml. It lives here, next to the validator it configures,
// so the tests that exercise the validator exercise the real policy — a
// hand-copied duplicate in a test would let the two drift apart silently.
//
// Matched case-insensitively: URI schemes and data-URI media types both are,
// and goldmark's own IsDangerousURL compares that way too. A case-sensitive
// prefix would see no "data:" in `datA:imAge/png;…` and pass it straight
// through to goldmark, which permits data:image/{png,gif,jpeg,webp,svg+xml}
// — so the raster URI this rule exists to reject would render.
//
// Leading C0 controls and spaces are trimmed for the same reason
// renderGnoLink trims them before its scheme check: they are not part of the
// URL a browser parses, but they DO shift the prefix, so `\tdata:image/png;…`
// would otherwise show this policy no "data:" at all and be waved through.
func AllowSvgDataImage(uri string) bool {
	const svgdata = "data:image/svg+xml"
	lower := strings.ToLower(trimLeadingControlAndSpace(uri))
	return !strings.HasPrefix(lower, "data:") || strings.HasPrefix(lower, svgdata)
}

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

		// Validate what the renderer will emit, not the raw bytes:
		// goldmark resolves entity references on the way out, so
		// `&#x64;ata:` reaches the browser as `data:` (resolveDestination).
		if !t.valFunc(string(resolveDestination(img.Destination))) {
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
