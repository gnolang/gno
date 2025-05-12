package gnoweb

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	md "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/yuin/goldmark"
	markdown "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type MarkdownRendererConfig struct {
	GoldmarkOptions []goldmark.Option
}

func NewDefaultMarkdownRendererConfig(chromaOptions []chromahtml.Option) *MarkdownRendererConfig {
	// Only allow svg data image
	allowSvgDataImage := func(uri string) bool {
		const svgdata = "image/svg+xml"
		return !strings.HasPrefix(uri, "data:") || strings.HasPrefix(uri, "data:"+svgdata)
	}

	goldmarkOptions := []goldmark.Option{
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithExtensions(
			markdown.NewHighlighting(
				markdown.WithFormatOptions(chromaOptions...),
			),

			extension.Strikethrough,
			extension.Table,

			md.NewGnoExtension(
				md.WithImageValidator(allowSvgDataImage),
			),
		),
	}

	return &MarkdownRendererConfig{
		GoldmarkOptions: goldmarkOptions,
	}
}

type MarkdownRenderer struct {
	logger   *slog.Logger
	markdown goldmark.Markdown
}

var _ ContentRenderer = (*MarkdownRenderer)(nil)

func NewMarkdownRenderer(logger *slog.Logger, cfg *MarkdownRendererConfig) *MarkdownRenderer {
	return &MarkdownRenderer{
		logger:   logger,
		markdown: goldmark.New(cfg.GoldmarkOptions...),
	}
}

func (mr *MarkdownRenderer) Render(w io.Writer, u *weburl.GnoURL, src []byte) (md.Toc, error) {
	ctx := md.NewGnoParserContext(u)

	// Use Goldmark for Markdown parsing
	doc := mr.markdown.Parser().Parse(text.NewReader(src), parser.WithContext(ctx))
	if err := mr.markdown.Renderer().Render(w, src, doc); err != nil {
		return md.Toc{}, fmt.Errorf("unable to render markdown at path %q: %w", u.Path, err)
	}

	toc, err := md.TocInspect(doc, src, md.TocOptions{MaxDepth: 6, MinDepth: 2})
	if err != nil {
		mr.logger.Warn("unable to inspect for TOC elements", "error", err)
	}

	return toc, nil
}
