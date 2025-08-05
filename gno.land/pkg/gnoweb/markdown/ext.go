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

var _ goldmark.Extender = (*GnoExtension)(nil)

type GnoExtension struct {
	cfg *config
}

// Option

type config struct {
	imgValidatorFunc ImageValidatorFunc
	extensions      []goldmark.Extender
}

type Option func(cfg *config)

func WithImageValidator(valFunc ImageValidatorFunc) Option {
	return func(cfg *config) {
		cfg.imgValidatorFunc = valFunc
	}
}

func WithExtensions(exts ...goldmark.Extender) Option {
	return func(cfg *config) {
		cfg.extensions = append(cfg.extensions, exts...)
	}
}

// newGnoExtension is a helper function to create Gno extensions with common logic
func newGnoExtension(defaultExtensions []goldmark.Extender, opts ...Option) *GnoExtension {
	cfg := &config{
		extensions: defaultExtensions,
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
		ExtColumns,   // Enable columns for realms
		ExtAlerts,    // Enable alerts for realms
		ExtLinks,     // Enable links for realms
		ExtForms,     // Enable forms for realms
		ExtMention,   // Enable mentions for realms
	}, opts...)
}

// NewDocumentationGnoExtension creates a Gno extension configured for documentation rendering
// Includes only ExtCodeExpand for clean, focused documentation
func NewDocumentationGnoExtension(opts ...Option) *GnoExtension {
	return newGnoExtension([]goldmark.Extender{
		ExtCodeExpand, // Only ExtCodeExpand for documentation
	}, opts...)
}

// Extend adds the Gno extension to the provided Goldmark markdown processor.
func (e *GnoExtension) Extend(m goldmark.Markdown) {
	// Add all configured extensions
	for _, ext := range e.cfg.extensions {
		ext.Extend(m)
	}

	// If set, setup images filter (ExtImageValidator has a different signature than other extensions)
	if e.cfg.imgValidatorFunc != nil {
		ExtImageValidator.Extend(m, e.cfg.imgValidatorFunc)
	}
}
