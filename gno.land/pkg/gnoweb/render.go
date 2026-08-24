package gnoweb

import (
	"bytes"
	"fmt"
	"html"
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

// maxMarkdownRenderBytes caps the markdown fed to goldmark per render. Beyond
// this size the residual super-linear goldmark vectors (reference-flood,
// nested-lists) cost too much CPU (~8 s at the 8 MiB RPC ceiling), so the
// content is shown as plain text rather than parsed. Legit realm output is
// well under 1 MiB (see maxRPCResponseSize commentary in client.go); this cap
// is distinct from and tighter than that transport guard.
const maxMarkdownRenderBytes = 1 << 20 // 1 MiB

// markdownPlainTextNotice precedes the plain-text fallback so a visitor
// understands why the page is unstyled. It mirrors the markup ExtAlerts emits
// for a `[!WARNING]` block (markdown/ext_alert.go) so it inherits gnoweb's
// alert styling and icon. Static, no attacker content.
const markdownPlainTextNotice = `<details class="gno-alert gno-alert-warning" open>
<summary>
<svg><use href="#ico-warning"></use></svg>Content too large to render<svg><use href="#ico-arrow"></use></svg>
</summary>
<div>
<p>This content exceeds the render limit and is shown as plain text below.</p>
</div>
</details>
`

// writeMarkdownPlainText handles input over maxMarkdownRenderBytes by emitting
// it as HTML-escaped plain text instead of feeding it to goldmark, whose
// super-linear parse paths are the cost being avoided. html.EscapeString keeps
// the (attacker-controlled) content from being reflected as live markup. It
// reports whether the fallback fired, so callers skip parsing and return their
// zero result.
func writeMarkdownPlainText(w io.Writer, src []byte) (handled bool, err error) {
	if len(src) <= maxMarkdownRenderBytes {
		return false, nil
	}
	if _, err := io.WriteString(w, markdownPlainTextNotice+`<pre class="gno-render-plaintext">`+html.EscapeString(string(src))+`</pre>`); err != nil {
		return true, fmt.Errorf("write plain-text fallback: %w", err)
	}
	return true, nil
}

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
			md.ExtEmphasis, // bound emphasis-parsing cost (yuin/goldmark#555)
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
	if handled, err := writeMarkdownPlainText(w, src); handled {
		return md.Toc{}, err
	}

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
	if handled, err := writeMarkdownPlainText(w, src); handled {
		return err
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
