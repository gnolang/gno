package gnoweb

// type MarkdownRendererConfig struct {
// 	GoldmarkOptions []goldmark.Option
// }

// type MarkdownRenderer struct {
// 	logger      *slog.Logger
// 	markdown    goldmark.Markdown
// 	Formatter   *chromahtml.Formatter
// 	chromaStyle *chroma.Style
// }

// var _ ContentRenderer = (*MarkdownRenderer)(nil)

// func NewMarkdownRenderer(logger *slog.Logger, cfg *MarkdownRendererConfig) *MarkdownRenderer {
// 	return &MarkdownRenderer{
// 		logger:   logger,
// 		markdown: goldmark.New(cfg.GoldmarkOptions...),
// 	}
// }

// func (mr *MarkdownRenderer) RenderMarkdown(w io.Writer, u *weburl.GnoURL, src []byte) (md.Toc, error) {
// 	ctx := md.NewGnoParserContext(u)

// 	// Use Goldmark for Markdown parsing
// 	doc := mr.markdown.Parser().Parse(text.NewReader(src), parser.WithContext(ctx))
// 	if err := mr.markdown.Renderer().Render(w, src, doc); err != nil {
// 		return md.Toc{}, fmt.Errorf("unable to render markdown at path %q: %w", u.Path, err)
// 	}

// 	toc, err := md.TocInspect(doc, src, md.TocOptions{MaxDepth: 6, MinDepth: 2})
// 	if err != nil {
// 		mr.logger.Warn("unable to inspect for TOC elements", "error", err)
// 	}

// 	return toc, nil
// }
