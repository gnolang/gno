package gnoweb

import (
	"fmt"
	"io"
	"log/slog"

	md "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/yuin/goldmark"
	markdown "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type MarkdownRenderer struct {
	logger   *slog.Logger
	goldmark goldmark.Markdown
}

func NewMarkdownRenderer(logger *slog.Logger, cfg *RenderConfig) *MarkdownRenderer {
	opts := append(cfg.GoldmarkOptions, goldmark.WithExtensions(
		markdown.NewHighlighting(
			markdown.WithFormatOptions(cfg.ChromaOptions...),
		),
	))
	return &MarkdownRenderer{logger, goldmark.New(opts...)}
}

func (mr *MarkdownRenderer) RenderRealm(w io.Writer, src []byte, u *weburl.GnoURL) (md.Toc, error) {
	ctx := md.NewGnoParserContext(u)

	// Use Goldmark for Markdown parsing
	doc := mr.goldmark.Parser().Parse(text.NewReader(src), parser.WithContext(ctx))
	if err := mr.goldmark.Renderer().Render(w, src, doc); err != nil {
		return md.Toc{}, fmt.Errorf("unable to render markdown at path %q: %w", u.Path, err)
	}

	toc, err := md.TocInspect(doc, src, md.TocOptions{MaxDepth: 6, MinDepth: 2})
	if err != nil {
		mr.logger.Warn("unable to inspect for TOC elements", "error", err)
	}

	return toc, nil
}
