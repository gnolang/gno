package gnoweb

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"io"
	"log/slog"
	gopath "path"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
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
// alert styling and icon. Static, no user-supplied content. A copy button leaves it
// out of what it copies.
const markdownPlainTextNotice = `<details class="gno-alert gno-alert-warning" data-copy-exclude open>
<summary>
<svg><use href="#ico-warning"></use></svg>Content too large to render<svg><use href="#ico-arrow"></use></svg>
</summary>
<div>
<p>This content exceeds the render limit and is shown as plain text below.</p>
</div>
</details>
`

// writePlainText writes src as HTML-escaped plain text, under
// markdownPlainTextNotice when notice is set. html.EscapeString keeps the
// (user-controlled) content from being reflected as live markup.
func writePlainText(w io.Writer, src []byte, notice bool) error {
	head := ""
	if notice {
		head = markdownPlainTextNotice
	}
	if _, err := io.WriteString(w, head+`<pre class="gno-render-plaintext">`+html.EscapeString(string(src))+`</pre>`); err != nil {
		return fmt.Errorf("write plain-text fallback: %w", err)
	}
	return nil
}

// maxRenderOutputBytes caps the HTML of one render. No markdown file in the tree
// renders past 276 KB, while the largest .gno, unicode/tables.gno, highlights
// from 210 KB to 4.2 MB, so a source file twice its size shows as plain text.
// 16 MiB would only double what one request may hold.
const maxRenderOutputBytes = 8 << 20 // 8 MiB

// errRenderOutputCap is the panic value cappedBuffer raises past its cap.
var errRenderOutputCap = errors.New("render output exceeds maxRenderOutputBytes")

// cappedBuffer panics past its cap, since only a panic stops goldmark's walk:
// past a write error it keeps highlighting, 2 to 3 times slower on code blocks.
// buf is not embedded, so bufio finds no WriteString.
type cappedBuffer struct {
	buf  bytes.Buffer
	left int
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	if len(p) > b.left {
		panic(errRenderOutputCap)
	}
	b.left -= len(p)
	return b.buf.Write(p)
}

// renderCapped copies what render writes to w, or writes nothing and reports
// true once render passes limit bytes. Other panics propagate.
func renderCapped(w io.Writer, limit int, render func(*cappedBuffer) error) (fellBack bool, err error) {
	buf := cappedBuffer{left: limit}
	defer func() {
		if r := recover(); r != nil {
			if r != errRenderOutputCap { //nolint:errorlint // sentinel panic value, compared by identity
				panic(r)
			}
			fellBack = true
		}
	}()
	if err := render(&buf); err != nil {
		return false, err
	}
	_, err = buf.buf.WriteTo(w)
	return false, err
}

// Renderer defines the interface for rendering realms, source files, and
// doc-context markdown (function/type/package documentation).
type Renderer interface {
	RenderRealm(w io.Writer, u *weburl.GnoURL, src []byte, ctx RealmRenderContext) (md.RealmMeta, error)
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
	// Links says which links of the document search engines may follow.
	Links md.LinkPolicy
	// File names the rendered file, a README, on the log line of a render past
	// maxRenderOutputBytes; empty for a realm's Render output.
	File string
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
	// NewRouter has already rejected an invalid gateway. Knowing neither the
	// chain domain nor the canonical origin, this only drops a direct
	// caller's malformed one, rather than rendering it as a relative URL.
	gateway, err := normalizeIPFSGateway(cfg.IPFSGateway, "", "")
	if err != nil {
		logger.Warn("ignoring invalid IPFS gateway", "error", err) // not the value: it may carry a secret
	}
	cfg.IPFSGateway = gateway

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
			md.ExtDocLinks,
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

// RenderRealm renders a realm to HTML and returns what the document says
// about itself: its table of contents, its heading and its summary.
func (r *HTMLRenderer) RenderRealm(w io.Writer, u *weburl.GnoURL, src []byte, ctx RealmRenderContext) (md.RealmMeta, error) {
	// Input past maxMarkdownRenderBytes never reaches goldmark, whose
	// super-linear parse paths are the cost being avoided.
	if len(src) > maxMarkdownRenderBytes {
		return md.RealmMeta{}, writePlainText(w, src, true)
	}

	var mdctx md.GnoContext
	mdctx.GnoURL = u
	mdctx.ChainId = ctx.ChainId
	mdctx.Remote = ctx.Remote
	mdctx.Domain = ctx.Domain
	mdctx.Links = ctx.Links
	mdctx.IPFSGateway = r.cfg.IPFSGateway
	mdctx.Mermaid = r.cfg.Mermaid

	pctx := md.NewGnoParserContext(mdctx)

	// Use Goldmark for Markdown parsing
	doc := r.gm.Parser().Parse(text.NewReader(src), parser.WithContext(pctx))
	var toc md.Toc
	fellBack, err := renderCapped(w, maxRenderOutputBytes, func(out *cappedBuffer) error {
		if err := r.gm.Renderer().Render(out, src, doc); err != nil {
			return err
		}
		var tocErr error
		if toc, tocErr = md.TocInspect(doc, src, md.TocOptions{MaxDepth: 6, MinDepth: 2}); tocErr != nil {
			r.logger.Warn("unable to inspect for TOC elements", "error", tocErr)
		}
		// The table of contents shows beside the article, so it is charged to
		// the same cap, then cut from the buffer.
		n := out.buf.Len()
		defer out.buf.Truncate(n)
		return components.NewTemplateComponent("ui/toc_realm", &components.RealmTOCData{Items: toc.Items}).Render(out)
	})
	if err != nil {
		return md.RealmMeta{}, fmt.Errorf("unable to render markdown at path %q: %w", u.Path, err)
	}
	if fellBack {
		r.logCapped(gopath.Join(u.Path, ctx.File), src)
		return md.RealmMeta{}, writePlainText(w, src, true)
	}

	meta := md.RealmMeta{Toc: toc}
	meta.Title, meta.Description = md.Lead(doc, src)
	return meta, nil
}

// RenderSource renders a source file into HTML with syntax highlighting based on its extension.
func (r *HTMLRenderer) RenderSource(w io.Writer, name string, src []byte) error {
	return newDocPage(r, name, 0).RenderSource(w, name, src)
}

func (r *HTMLRenderer) formatSource(w io.Writer, name string, src []byte) error {
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
	return newDocPage(r, "", 0).RenderDocumentation(w, src)
}

// pageDocRenderer renders the doc comments, signatures and cards of one page
// against one maxRenderOutputBytes budget. It is built per request and used
// sequentially.
type pageDocRenderer struct {
	*HTMLRenderer
	path    string // logged when the budget runs out
	left    int    // bytes the page may still write
	spent   bool   // later doc comments and signatures are plain text
	noticed bool
}

// newPageDocRenderer returns r with the page budget left after used bytes, or
// r itself when it is not an *HTMLRenderer.
func newPageDocRenderer(r Renderer, path string, used int) Renderer {
	base, ok := r.(*HTMLRenderer)
	if !ok {
		return r
	}
	return newDocPage(base, path, used)
}

func newDocPage(r *HTMLRenderer, path string, used int) *pageDocRenderer {
	return &pageDocRenderer{HTMLRenderer: r, path: path, left: maxRenderOutputBytes - used}
}

func (p *pageDocRenderer) RenderDocumentation(w io.Writer, src []byte) error {
	if len(src) == 0 {
		return nil
	}
	if len(src) > maxMarkdownRenderBytes {
		return p.writePlain(w, src)
	}
	return p.render(w, src, func(out *cappedBuffer) error {
		return p.documentationGM.Convert(src, out)
	})
}

func (p *pageDocRenderer) RenderSource(w io.Writer, name string, src []byte) error {
	return p.render(w, src, func(out *cappedBuffer) error {
		return p.formatSource(out, name, src)
	})
}

// render writes what fn renders, charged to the page. A render past what is left
// spends the rest, so no later one renders up to the remainder only to fall back;
// one finding nothing left, a README having used it, spends it too. Either logs.
func (p *pageDocRenderer) render(w io.Writer, src []byte, fn func(*cappedBuffer) error) error {
	if !p.spent {
		if p.left > 0 {
			cw := countingWriter{w: w}
			fellBack, err := renderCapped(&cw, p.left, fn)
			p.left -= cw.n
			if !fellBack {
				return err
			}
		}
		p.spent = true
		p.logCapped(p.path, src)
	}
	return p.writePlain(w, src)
}

// renderPart writes what fn renders for a part of the page that is not
// markdown, a template card, charged to r's page budget when r has one. held is
// what the part embeds that the page was already charged for, and beside what
// the part adds elsewhere on the page, charged with it. It reports false,
// writing nothing, when the part passes what is left.
func renderPart(r Renderer, w io.Writer, held, beside int, fn func(io.Writer) error) (bool, error) {
	p, ok := r.(*pageDocRenderer)
	if !ok {
		return true, fn(w)
	}
	cw := countingWriter{w: w}
	fellBack, err := renderCapped(&cw, p.left+held-beside, func(out *cappedBuffer) error { return fn(out) })
	if fellBack {
		p.logger.Warn("render output past maxRenderOutputBytes, part left out", "path", p.path)
	}
	if fellBack || err != nil {
		return false, err
	}
	p.left -= cw.n - held + beside
	return true, nil
}

// writePlain writes src as plain text charged to the page, under the notice the
// first time the page's doc comments and signatures fall back.
func (p *pageDocRenderer) writePlain(w io.Writer, src []byte) error {
	cw := countingWriter{w: w}
	err := writePlainText(&cw, src, !p.noticed)
	p.noticed = true
	p.left -= cw.n
	return err
}

// logCapped records a render that passed maxRenderOutputBytes, so how often real
// pages reach the cap can be read off the logs.
func (r *HTMLRenderer) logCapped(path string, src []byte) {
	r.logger.Warn("render output past maxRenderOutputBytes, serving plain text", "path", path, "len", len(src))
}

type countingWriter struct {
	w io.Writer
	n int
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += n
	return n, err
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
		var dark bytes.Buffer
		if err := r.ch.WriteCSS(&dark, r.cfg.ChromaDarkStyle); err != nil {
			return fmt.Errorf("writing dark chroma CSS: %w", err)
		}
		// The highlighted HTML always carries the light style's mode class,
		// so the dark rules must target it too, or [data-theme="dark"] keeps
		// the light palette (unreadable on a dark background). The wrapper
		// and background rules are dropped: gnoweb paints the code frames.
		for _, line := range bytes.SplitAfter(dark.Bytes(), []byte("\n")) {
			if bytes.HasPrefix(line, []byte("/* PreWrapper */")) || bytes.HasPrefix(line, []byte("/* Background */")) {
				continue
			}
			buf.Write(bytes.ReplaceAll(line, []byte(".chroma-dark"), []byte(".chroma-light")))
		}
		buf.WriteString("}\n")
		if _, err := buf.WriteTo(w); err != nil {
			return fmt.Errorf("writing dark chroma CSS: %w", err)
		}
	}

	return nil
}
