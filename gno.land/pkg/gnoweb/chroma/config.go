package chroma

import (
	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	markdown "github.com/yuin/goldmark-highlighting/v2"
)

const defaultStyleName = "friendly"

var DefaultChromaRenderStyle = styles.Get(defaultStyleName)

// Default options for Chroma syntax highlighting
var defaultChromaOptions = []chromahtml.Option{
	chromahtml.WithLineNumbers(true),
	chromahtml.WithLinkableLineNumbers(true, "L"),
	chromahtml.WithClasses(true),
	chromahtml.ClassPrefix("chroma-"),
}

// Shared Chroma instance for consistent syntax highlighting across all components
var (
	sharedChromaFormatter *chromahtml.Formatter
	sharedChromaStyle     *chroma.Style
)

func init() {
	// Initialize shared Chroma components
	sharedChromaFormatter = chromahtml.New(defaultChromaOptions...)
	sharedChromaStyle = DefaultChromaRenderStyle
}

// GetSharedChromaComponents returns the shared Chroma components
func GetSharedChromaComponents() (*chromahtml.Formatter, *chroma.Style) {
	return sharedChromaFormatter, sharedChromaStyle
}

// NewDefaultChromaOptions returns the default Chroma options for syntax highlighting.
func NewDefaultChromaOptions() []chromahtml.Option {
	return defaultChromaOptions
}

// NewHighlightingExtension returns a configured highlighting extension
func NewHighlightingExtension() goldmark.Extender {
	return markdown.NewHighlighting(
		markdown.WithStyle(defaultStyleName),
		markdown.WithFormatOptions(defaultChromaOptions...),
	)
}
