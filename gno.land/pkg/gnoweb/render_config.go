package gnoweb

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	md "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/yuin/goldmark"
	markdown "github.com/yuin/goldmark-highlighting/v2"
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
	cfg.GoldmarkOptions = NewDocumentationGoldmarkOptions() // Use documentation config by default
	cfg.ChromaOptions = NewDefaultChromaOptions()
	return cfg
}

// NewRealmGoldmarkOptions returns Goldmark options for realm rendering
// Includes all realm-specific features
func NewRealmGoldmarkOptions() []goldmark.Option {
	// Only allow svg data image
	allowSvgDataImage := func(uri string) bool {
		const svgdata = "image/svg+xml"
		return !strings.HasPrefix(uri, "data:") || strings.HasPrefix(uri, "data:"+svgdata)
	}

	return []goldmark.Option{
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithExtensions(
			extension.Strikethrough,
			extension.Table,
			extension.Footnote,
			extension.TaskList,
			// Realm-specific Gno extension with all features
			md.NewRealmGnoExtension(
				md.WithImageValidator(allowSvgDataImage),
			),
			markdown.NewHighlighting(
				markdown.WithStyle("friendly"),
				markdown.WithFormatOptions(NewDefaultChromaOptions()...),
			),
		),
	}
}

// NewDocumentationGoldmarkOptions returns Goldmark options for documentation rendering
// Includes only ExtCodeExpand for clean, focused documentation
func NewDocumentationGoldmarkOptions() []goldmark.Option {
	return []goldmark.Option{
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithExtensions(
			// Documentation-specific Gno extension (only ExtCodeExpand)
			md.NewDocumentationGnoExtension(),
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
