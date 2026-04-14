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
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/safeurl"
	"github.com/yuin/goldmark"
)

var _ goldmark.Extender = (*GnoExtension)(nil)

type GnoExtension struct {
	cfg *config
}

// Option

type config struct {
	imgValidatorFunc ImageValidatorFunc
	safeURLValidator *safeurl.Validator
}

type Option func(cfg *config)

func WithImageValidator(valFunc ImageValidatorFunc) Option {
	return func(cfg *config) {
		cfg.imgValidatorFunc = valFunc
	}
}

// WithSafeURLValidator sets the SafeURL validator for URL safety checking.
func WithSafeURLValidator(v *safeurl.Validator) Option {
	return func(cfg *config) {
		cfg.safeURLValidator = v
	}
}

func NewGnoExtension(opts ...Option) *GnoExtension {
	var cfg config
	for _, opt := range opts {
		opt(&cfg)
	}

	return &GnoExtension{&cfg}
}

// Extend adds the Gno extension to the provided Goldmark markdown processor.
func (e *GnoExtension) Extend(m goldmark.Markdown) {
	// Add column extension
	ExtColumns.Extend(m)

	// Add alert extension
	ExtAlerts.Extend(m)

	// Add SafeURL extension (must run before link extension at priority 400)
	if e.cfg.safeURLValidator != nil {
		ExtSafeURL.Extend(m, e.cfg.safeURLValidator)
	}

	// Add link extension (runs at priority 500)
	ExtLinks.Extend(m)

	// Add form / inputs extension
	ExtForms.Extend(m)

	// Add mentions extension
	ExtMention.Extend(m)

	// If set, setup images filter
	if e.cfg.imgValidatorFunc != nil {
		ExtImageValidator.Extend(m, e.cfg.imgValidatorFunc)
	}
}
