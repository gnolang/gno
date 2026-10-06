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
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"
)

var _ goldmark.Extender = (*GnoExtension)(nil)

type GnoExtension struct {
	cfg *config
}

// Option

type config struct {
	imgValidatorFunc ImageValidatorFunc
	peerBlockParsers []util.PrioritizedValue
}

type Option func(cfg *config)

func WithImageValidator(valFunc ImageValidatorFunc) Option {
	return func(cfg *config) {
		cfg.imgValidatorFunc = valFunc
	}
}

// WithPeerBlockParsers declares the block parsers that the goldmark
// extensions loaded alongside the Gno extension register, such as
// extension.Footnote's. Display math ends at any line one of them would
// open to interrupt a paragraph, as it does at the CommonMark blocks and
// gnoweb's own: goldmark cannot list the parsers an extension registered
// before this one, so they must be passed here.
func WithPeerBlockParsers(bps ...parser.BlockParser) Option {
	return func(cfg *config) {
		for _, bp := range bps {
			// The priority is unused: the parsers are only probed.
			cfg.peerBlockParsers = append(cfg.peerBlockParsers, util.Prioritized(bp, 0))
		}
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
	// Record the block parsers the extensions below register: display math
	// must end where any of them would interrupt a paragraph.
	rec := &blockParserRecorder{Parser: m.Parser(), cfg: parser.NewConfig()}
	m.SetParser(rec)

	// Bound goldmark emphasis-parsing cost (yuin/goldmark#555) before anything
	// else parses attacker-controlled markdown.
	ExtEmphasis.Extend(m)

	// Add foreign extension. Image validator is forwarded so the
	// inner instance inherits it. gno-form is intentionally NOT
	// loaded inside the foreign sandbox (forms are interactive UI
	// and never permitted in foreign-controlled bytes).
	ExtForeign.Extend(m, e.cfg.imgValidatorFunc)

	// Add column extension
	ExtColumns.Extend(m)

	// Add alert extension
	ExtAlerts.Extend(m)

	// Add link extension
	ExtLinks.Extend(m)

	// Add form / inputs extension
	ExtForms.Extend(m)

	// Add mentions extension
	ExtMention.Extend(m)

	m.SetParser(rec.Parser)

	// Add math extension
	NewExtMath(append(rec.cfg.BlockParsers, e.cfg.peerBlockParsers...)...).Extend(m)

	// If set, setup images filter
	if e.cfg.imgValidatorFunc != nil {
		ExtImageValidator.Extend(m, e.cfg.imgValidatorFunc)
	}
}

// blockParserRecorder forwards parser options to Parser and records them in
// cfg, where the block parsers they add can be read back.
type blockParserRecorder struct {
	parser.Parser
	cfg *parser.Config
}

func (r *blockParserRecorder) AddOptions(opts ...parser.Option) {
	for _, opt := range opts {
		opt.SetParserOption(r.cfg)
	}
	r.Parser.AddOptions(opts...)
}
