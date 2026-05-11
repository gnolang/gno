package gnoweb

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	gopath "path"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	md "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/yuin/goldmark"
	markdown "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Renderer defines the interface for rendering realms, source files, and
// doc-context markdown (function/type/package documentation).
type Renderer interface {
	RenderRealm(w io.Writer, u *weburl.GnoURL, src []byte, ctx RealmRenderContext) (md.Toc, error)
	RenderSource(w io.Writer, name string, src []byte) error
	// RenderDocumentation renders doc-context markdown (from vm/qdoc) to HTML.
	// Fenced and indented code blocks are wrapped in collapsible <details>
	// with Chroma highlighting. HTML escaping is delegated to Goldmark.
	RenderDocumentation(w io.Writer, src []byte) error
}

// RealmRenderContext holds context information for rendering realms
type RealmRenderContext struct {
	ChainId string
	Remote  string
	Domain  string
}

// HTMLRenderer implements the Renderer interface for HTML output.
type HTMLRenderer struct {
	logger *slog.Logger
	cfg    *RenderConfig
	client ClientAdapter

	gm              goldmark.Markdown // realm context
	documentationGM goldmark.Markdown // doc context
	ch              *chromahtml.Formatter
}

func NewHTMLRenderer(logger *slog.Logger, cfg RenderConfig, client ClientAdapter) *HTMLRenderer {
	gmOpts := append(cfg.GoldmarkOptions, goldmark.WithExtensions(
		markdown.NewHighlighting(
			markdown.WithFormatOptions(cfg.ChromaOptions...), // force using chroma config
		),
	))

	// Doc-context renderer. parser.WithAttribute consumes the Pandoc-style
	// `{#id}` heading suffixes emitted by gnovm/pkg/doc; ExtCodeExpand wraps
	// fenced code blocks in collapsible <details>.
	docFormatter := chromahtml.New(cfg.ChromaOptions...)
	docOpts := []goldmark.Option{
		goldmark.WithExtensions(
			markdown.NewHighlighting(markdown.WithFormatOptions(cfg.ChromaOptions...)),
			md.ExtCodeExpand(docFormatter, cfg.ChromaStyle),
		),
		goldmark.WithParserOptions(parser.WithAttribute()),
	}

	return &HTMLRenderer{
		logger:          logger,
		cfg:             &cfg,
		client:          client,
		gm:              goldmark.New(gmOpts...),
		documentationGM: goldmark.New(docOpts...),
		ch:              chromahtml.New(cfg.ChromaOptions...),
	}
}

// RenderRealm renders a realm to HTML and returns a table of contents.
func (r *HTMLRenderer) RenderRealm(w io.Writer, u *weburl.GnoURL, src []byte, ctx RealmRenderContext) (md.Toc, error) {
	var mdctx md.GnoContext
	mdctx.GnoURL = u
	mdctx.ChainId = ctx.ChainId
	mdctx.Remote = ctx.Remote
	mdctx.Domain = ctx.Domain

	pctx := md.NewGnoParserContext(mdctx)

	// Use Goldmark for Markdown parsing
	doc := r.gm.Parser().Parse(text.NewReader(src), parser.WithContext(pctx))
	if err := r.gm.Renderer().Render(w, src, doc); err != nil {
		return md.Toc{}, fmt.Errorf("unable to render markdown at path %q: %w", u.Path, err)
	}

	toc, err := md.TocInspect(doc, src, md.TocOptions{MaxDepth: 6, MinDepth: 2})
	if err != nil {
		r.logger.Warn("unable to inspect for TOC elements", "error", err)
	}

	return toc, nil
}

// RenderSource renders a source file into HTML with syntax highlighting based on its extension.
func (r *HTMLRenderer) RenderSource(w io.Writer, name string, src []byte) error {
	var lexer chroma.Lexer

	// Determine the lexer to be used based on the file extension.
	switch strings.ToLower(gopath.Ext(name)) {
	case ".gno":
		lexer = lexers.Get("go")
	case ".md":
		lexer = lexers.Get("markdown")
	case ".mod":
		lexer = lexers.Get("gomod")
	case ".toml":
		lexer = lexers.Get("toml")
	default:
		lexer = lexers.Get("txt") // Unsupported file type, default to plain text.
	}

	if lexer == nil {
		return fmt.Errorf("unsupported lexer for file %q", name)
	}

	iterator, err := lexer.Tokenise(nil, string(src))
	if err != nil {
		return fmt.Errorf("unable to tokenise %q: %w", name, err)
	}

	if err := r.ch.Format(w, r.cfg.ChromaStyle, iterator); err != nil {
		return fmt.Errorf("unable to format source file %q: %w", name, err)
	}

	return nil
}

// RenderDocumentation writes the HTML representation of doc-context markdown
// to w. HTML escaping is delegated to Goldmark.
func (r *HTMLRenderer) RenderDocumentation(w io.Writer, src []byte) error {
	if len(src) == 0 {
		return nil
	}
	if err := r.documentationGM.Convert(src, w); err != nil {
		return fmt.Errorf("render documentation: %w", err)
	}
	return nil
}

// WriteChromaCSS writes the CSS for syntax highlighting to the provided writer.
// It outputs the light theme by default and, if configured, the dark theme
// scoped under [data-theme="dark"] using CSS nesting.
func (r *HTMLRenderer) WriteChromaCSS(w io.Writer) error {
	if err := r.ch.WriteCSS(w, r.cfg.ChromaStyle); err != nil {
		return fmt.Errorf("writing light chroma CSS: %w", err)
	}

	if r.cfg.ChromaDarkStyle != nil {
		var buf bytes.Buffer
		buf.WriteString("\n[data-theme=\"dark\"] {\n")
		if err := r.ch.WriteCSS(&buf, r.cfg.ChromaDarkStyle); err != nil {
			return fmt.Errorf("writing dark chroma CSS: %w", err)
		}
		buf.WriteString("}\n")
		if _, err := buf.WriteTo(w); err != nil {
			return fmt.Errorf("writing dark chroma CSS: %w", err)
		}
	}

	return nil
}
