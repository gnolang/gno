package gnoweb

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	gopath "path"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	md "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/yuin/goldmark"
	markdown "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Renderer defines the interface for rendering realms and source files.
type Renderer interface {
	RenderRealm(w io.Writer, u *weburl.GnoURL, src []byte, ctx RealmRenderContext) (md.Toc, error)
	RenderSource(w io.Writer, name string, src []byte) error
}

// RealmRenderContext holds context information for rendering realms
type RealmRenderContext struct {
	ChainId string
	Remote  string
}

// HTMLRenderer implements the Renderer interface for HTML output.
type HTMLRenderer struct {
	logger *slog.Logger
	cfg    *RenderConfig
	client ClientAdapter

	gm goldmark.Markdown
	ch *chromahtml.Formatter
}

func NewHTMLRenderer(logger *slog.Logger, cfg RenderConfig, client ClientAdapter) *HTMLRenderer {
	gmOpts := append(cfg.GoldmarkOptions, goldmark.WithExtensions(
		markdown.NewHighlighting(
			markdown.WithFormatOptions(cfg.ChromaOptions...), // force using chroma config
		),
	))
	return &HTMLRenderer{
		logger: logger,
		cfg:    &cfg,
		client: client,
		gm:     goldmark.New(gmOpts...),
		ch:     chromahtml.New(cfg.ChromaOptions...),
	}
}

// RenderRealm renders a realm to HTML and returns a table of contents.
func (r *HTMLRenderer) RenderRealm(w io.Writer, u *weburl.GnoURL, src []byte, ctx RealmRenderContext) (md.Toc, error) {
	var mdctx md.GnoContext
	mdctx.GnoURL = u
	mdctx.ChainId = ctx.ChainId
	mdctx.Remote = ctx.Remote

	var once sync.Once

	// Create a lazy function to get funcs signature
	mdctx.RealmFuncSigGetter = md.RealmFuncSigGetter(func(fn string) (*vm.FunctionSignature, error) {
		var msigs map[string]*vm.FunctionSignature

		var err error
		once.Do(func() {
			if r.client == nil {
				r.logger.Warn("no client configured for fetching function signatures")
				return
			}

			ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
			defer cancel()

			var sigs vm.FunctionSignatures
			sigs, err = r.client.ListFuncs(ctx, u.Path)
			if err != nil {
				r.logger.Error("unable to fetch func signature lists", "path", u.Path, "err", err)
				return
			}

			msigs = make(map[string]*vm.FunctionSignature)
			for _, sig := range sigs {
				msigs[sig.FuncName] = &sig
			}

		})
		if err != nil {
			return nil, err
		}

		return msigs[fn], err
	})

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

// WriteChromaCSS writes the CSS for syntax highlighting to the provided writer.
func (r *HTMLRenderer) WriteChromaCSS(w io.Writer) error {
	return r.ch.WriteCSS(w, r.cfg.ChromaStyle)
}
