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
	enableCodeExpand bool
	enableColumns    bool
	enableAlerts     bool
	enableLinks      bool
	enableForms      bool
	enableMentions   bool
}

type Option func(cfg *config)

func WithImageValidator(valFunc ImageValidatorFunc) Option {
	return func(cfg *config) {
		cfg.imgValidatorFunc = valFunc
	}
}

func WithCodeExpand(enable bool) Option {
	return func(cfg *config) {
		cfg.enableCodeExpand = enable
	}
}

func WithColumns(enable bool) Option {
	return func(cfg *config) {
		cfg.enableColumns = enable
	}
}

func WithAlerts(enable bool) Option {
	return func(cfg *config) {
		cfg.enableAlerts = enable
	}
}

func WithLinks(enable bool) Option {
	return func(cfg *config) {
		cfg.enableLinks = enable
	}
}

func WithForms(enable bool) Option {
	return func(cfg *config) {
		cfg.enableForms = enable
	}
}

func WithMentions(enable bool) Option {
	return func(cfg *config) {
		cfg.enableMentions = enable
	}
}

// NewGnoExtension creates a new Gno extension with minimal default configuration.
func NewGnoExtension(opts ...Option) *GnoExtension {
	cfg := &config{
		// Minimal default configuration - only essential extensions
		enableCodeExpand: false, // Disabled by default, enable only where needed
		enableColumns:    false, // Disabled by default
		enableAlerts:     false, // Disabled by default
		enableLinks:      true,  // Essential for most use cases
		enableForms:      false, // Disabled by default
		enableMentions:   false, // Disabled by default
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return &GnoExtension{cfg}
}

// Extend adds the Gno extension to the provided Goldmark markdown processor.
func (e *GnoExtension) Extend(m goldmark.Markdown) {
	// Add column extension
	if e.cfg.enableColumns {
		ExtColumns.Extend(m)
	}

	// Add alert extension
	if e.cfg.enableAlerts {
		ExtAlerts.Extend(m)
	}

	// Add link extension
	if e.cfg.enableLinks {
		ExtLinks.Extend(m)
	}

	// Add form / inputs extension
	if e.cfg.enableForms {
		ExtForms.Extend(m)
	}

	// Add mentions extension
	if e.cfg.enableMentions {
		ExtMention.Extend(m)
	}

	// Add expandable code blocks extension
	if e.cfg.enableCodeExpand {
		ExtCodeExpand.Extend(m)
	}

	// If set, setup images filter
	if e.cfg.imgValidatorFunc != nil {
		ExtImageValidator.Extend(m, e.cfg.imgValidatorFunc)
	}
}
