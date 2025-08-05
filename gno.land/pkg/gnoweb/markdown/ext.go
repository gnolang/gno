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
	extdoc "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown/ext_doc"
	extrealm "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown/ext_realm"
	extshared "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown/ext_shared"
)

// ImageValidatorFunc validates image URLs. It should return `true` for any valid image URL.
type ImageValidatorFunc = extrealm.ImageValidatorFunc

var _ goldmark.Extender = (*GnoExtension)(nil)

type GnoExtension struct {
	cfg *config
}

// Option

type config struct {
	imgValidatorFunc ImageValidatorFunc
	extensions       []goldmark.Extender
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
		extrealm.ExtColumns, // Enable columns for realms
		extrealm.ExtAlerts,  // Enable alerts for realms
		extshared.ExtLinks,  // Enable links for realms
		extrealm.ExtForms,   // Enable forms for realms
		extrealm.ExtMention, // Enable mentions for realms
	}, opts...)
}

// NewDocumentationGnoExtension creates a Gno extension configured for documentation rendering
// Includes ExtCodeExpand and ExtLinks for clean, focused documentation
func NewDocumentationGnoExtension(opts ...Option) *GnoExtension {
	return newGnoExtension([]goldmark.Extender{
		extdoc.ExtCodeExpand, // Expandable code blocks for documentation
		extshared.ExtLinks,   // Enable links for documentation
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
		extrealm.ExtImageValidator.Extend(m, e.cfg.imgValidatorFunc)
	}
}
