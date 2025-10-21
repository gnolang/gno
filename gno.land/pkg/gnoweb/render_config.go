package gnoweb

import (
	"strings"

	chromaconfig "github.com/gnolang/gno/gno.land/pkg/gnoweb/chroma"
	md "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

// RenderConfig holds configuration for Markdown rendering.
type RenderConfig struct {
	GoldmarkOptions []goldmark.Option
}

// NewDefaultRenderConfig returns a RenderConfig with default options.
func NewDefaultRenderConfig() (cfg RenderConfig) {
	cfg.GoldmarkOptions = NewDocumentationGoldmarkOptions() // Use documentation config by default
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
			// Use centralized highlighting extension
			chromaconfig.NewHighlightingExtension(),
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
