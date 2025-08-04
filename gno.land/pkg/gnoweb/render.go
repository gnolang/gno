package gnoweb

import (
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
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Renderer defines the interface for rendering realms and source files.
type Renderer interface {
	RenderRealm(w io.Writer, u *weburl.GnoURL, src []byte) (md.Toc, error)
	RenderSource(w io.Writer, name string, src []byte) error
	RenderDocumentation(w io.Writer, u *weburl.GnoURL, src []byte) error
}

// HTMLRenderer implements the Renderer interface for HTML output.
type HTMLRenderer struct {
	logger *slog.Logger
	cfg    *RenderConfig

	// Separate Goldmark instances for different contexts
	realmGM         goldmark.Markdown // For realm rendering
	documentationGM goldmark.Markdown // For documentation rendering
	ch              *chromahtml.Formatter
}

func NewHTMLRenderer(logger *slog.Logger, cfg RenderConfig) *HTMLRenderer {
	return &HTMLRenderer{
		logger: logger,
		cfg:    &cfg,
		// Create dedicated instances for each context
		realmGM:         goldmark.New(NewRealmGoldmarkOptions()...),
		documentationGM: goldmark.New(NewDocumentationGoldmarkOptions()...),
		ch:              chromahtml.New(cfg.ChromaOptions...),
	}
}

// RenderRealm renders a realm to HTML and returns a table of contents.
func (r *HTMLRenderer) RenderRealm(w io.Writer, u *weburl.GnoURL, src []byte) (md.Toc, error) {
	// Use the dedicated realm Goldmark instance (pre-created for performance)
	ctx := md.NewGnoParserContext(u)

	// Use Goldmark for Markdown parsing
	doc := r.realmGM.Parser().Parse(text.NewReader(src), parser.WithContext(ctx))
	if err := r.realmGM.Renderer().Render(w, src, doc); err != nil {
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

// RenderDocumentation renders documentation with expandable code blocks enabled
func (r *HTMLRenderer) RenderDocumentation(w io.Writer, u *weburl.GnoURL, src []byte) error {
	// Use the dedicated documentation Goldmark instance
	ctx := md.NewGnoParserContext(u)

	// Parse and render the markdown with context
	doc := r.documentationGM.Parser().Parse(text.NewReader(src), parser.WithContext(ctx))
	if err := r.documentationGM.Renderer().Render(w, src, doc); err != nil {
		return fmt.Errorf("unable to render documentation: %w", err)
	}

	return nil
}

// WriteChromaCSS writes the CSS for syntax highlighting to the provided writer.
func (r *HTMLRenderer) WriteChromaCSS(w io.Writer) error {
	return r.ch.WriteCSS(w, r.cfg.ChromaStyle)
}
