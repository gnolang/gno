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

type RenderConfig struct {
	ChromaStyle     *chroma.Style
	ChromaOptions   []chromahtml.Option
	GoldmarkOptions []goldmark.Option
}

func NewDefaultRenderConfig() (cfg RenderConfig) {
	cfg.ChromaStyle = DefaultChromaRenderStyle
	cfg.GoldmarkOptions = NewDefaultGoldmarkOptions()
	cfg.ChromaOptions = NewDefaultChromaOptions()
	return cfg
}

func NewDefaultGoldmarkOptions() []goldmark.Option {
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
			md.NewGnoExtension(
				md.WithImageValidator(allowSvgDataImage),
			),
		),
	}
}

func NewDefaultChromaOptions() []chromahtml.Option {
	return []chromahtml.Option{
		chromahtml.WithLineNumbers(true),
		chromahtml.WithLinkableLineNumbers(true, "L"),
		chromahtml.WithClasses(true),
		chromahtml.ClassPrefix("chroma-"),
	}
}
