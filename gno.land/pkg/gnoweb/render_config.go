package gnoweb

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	md "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

var DefaultChromaRenderStyle = styles.Get("friendly")

// RenderConfig holds configuration for syntax highlighting and Markdown rendering.
type RenderConfig struct {
	ChromaStyle     *chroma.Style
	ChromaOptions   []chromahtml.Option
	GoldmarkOptions []goldmark.Option
}

// NewDefaultRenderConfig returns a RenderConfig with default styles and options.
func NewDefaultRenderConfig() (cfg RenderConfig) {
	cfg.ChromaStyle = DefaultChromaRenderStyle
	cfg.GoldmarkOptions = NewDefaultGoldmarkOptions()
	cfg.ChromaOptions = NewDefaultChromaOptions()
	return cfg
}

// NewDefaultGoldmarkOptions returns the default Goldmark options for Markdown rendering.
func NewDefaultGoldmarkOptions() []goldmark.Option {
	// Only allow svg data image
	allowSvgDataImage := func(uri string) bool {
		const svgdata = "image/svg+xml"
		return !strings.HasPrefix(uri, "data:") || strings.HasPrefix(uri, "data:"+svgdata)
	}

	var opts []md.Option
	opts = append(opts, md.WithImageValidator(allowSvgDataImage))

	opts = append(opts, md.WithContentFilter(md.DefaultContentFilter))

	return []goldmark.Option{
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithExtensions(
			extension.Strikethrough,
			extension.Table,
			extension.Footnote,
			extension.TaskList,
			md.NewGnoExtension(opts...),
		),
	}
}

// NewDefaultChromaOptions returns the default Chroma options for syntax highlighting.
func NewDefaultChromaOptions() []chromahtml.Option {
	return []chromahtml.Option{
		chromahtml.WithLineNumbers(true),
		chromahtml.WithLinkableLineNumbers(true, "L"),
		chromahtml.WithClasses(true),
		chromahtml.ClassPrefix("chroma-"),
	}
}
