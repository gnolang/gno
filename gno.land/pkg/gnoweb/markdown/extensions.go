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

	// Import real extensions from subdirectories
	exts "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown/extensions"
	extsdoc "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown/extensions/doc"
)

// ImageValidatorFunc validates image URLs. It should return `true` for any valid image URL.
type ImageValidatorFunc = exts.ImageValidatorFunc

var _ goldmark.Extender = (*GnoExtension)(nil)

type config struct {
	imgValidatorFunc ImageValidatorFunc
	extensions       []goldmark.Extender
}

type GnoExtension struct {
	cfg *config
}

// Extend adds the Gno extension to the provided Goldmark markdown processor.
func (e *GnoExtension) Extend(m goldmark.Markdown) {
	// Add all configured extensions
	for _, ext := range e.cfg.extensions {
		ext.Extend(m)
	}

	// If set, setup images filter (ExtImageValidator has a different signature than other extensions)
	if e.cfg.imgValidatorFunc != nil {
		exts.ExtImageValidator.Extend(m, e.cfg.imgValidatorFunc)
	}
}

// Option

type Option func(cfg *config)

func WithImageValidator(valFunc ImageValidatorFunc) Option {
	return func(cfg *config) {
		cfg.imgValidatorFunc = valFunc
	}
}

// newGnoExtension is a helper function to create Gno extensions with common logic
func newGnoExtension(exts []goldmark.Extender, opts ...Option) *GnoExtension {
	cfg := &config{
		extensions: exts,
	}

	// Apply all options
	for _, opt := range opts {
		opt(cfg)
	}

	return &GnoExtension{cfg}
}

// NewRealmGnoExtension creates a Gno extension configured for realm rendering
// Includes all realm-specific features with full markdown support
func NewRealmGnoExtension(opts ...Option) *GnoExtension {
	return newGnoExtension([]goldmark.Extender{
		exts.ExtColumns, // Enable columns for realms
		exts.ExtAlerts,  // Enable alerts for realms
		exts.ExtLinks,   // Enable links for realms
		exts.ExtForms,   // Enable forms for realms
		exts.ExtMention, // Enable mentions for realms
	}, opts...)
}

// NewDocumentationGnoExtension creates a Gno extension configured for documentation rendering
// Includes ExtCodeExpand and ExtLinks for clean, focused documentation
func NewDocumentationGnoExtension(opts ...Option) *GnoExtension {
	return newGnoExtension([]goldmark.Extender{
		extsdoc.ExtCodeExpand, // Expandable code blocks for documentation
		exts.ExtLinks,         // Enable links for documentation
	}, opts...)
}
