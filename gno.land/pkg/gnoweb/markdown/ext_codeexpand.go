package markdown

import (
	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// defaultLexer is a safe text-passthrough lexer used when no language is
// specified or the requested language is not recognised. It emits the raw
// content as a single token, so Chroma still escapes HTML entities but does
// not attempt syntactic tokenisation.
var defaultLexer = lexers.Fallback

// ExtCodeExpand returns a Goldmark extension that renders fenced and indented
// code blocks as collapsible <details class="doc-example"> disclosures with
// Chroma syntax highlighting applied to the code content.
//
// The formatter and style are dependency-injected so the caller controls
// highlighting consistency with the surrounding renderer. No global state.
func ExtCodeExpand(formatter *chromahtml.Formatter, style *chroma.Style) goldmark.Extender {
	return &codeExpandExtension{formatter: formatter, style: style}
}

type codeExpandExtension struct {
	formatter *chromahtml.Formatter
	style     *chroma.Style
}

func (e *codeExpandExtension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&codeExpandRenderer{formatter: e.formatter, style: e.style}, 0),
	))
}

type codeExpandRenderer struct {
	formatter *chromahtml.Formatter
	style     *chroma.Style
}

func (r *codeExpandRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.render)
	reg.Register(ast.KindCodeBlock, r.render)
}

func (r *codeExpandRenderer) render(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	// Collect the code content from the block's line segments.
	var lines *text.Segments
	var language []byte
	switch block := n.(type) {
	case *ast.FencedCodeBlock:
		language = block.Language(source)
		lines = block.Lines()
	case *ast.CodeBlock:
		lines = block.Lines()
	default:
		return ast.WalkContinue, nil
	}

	var code []byte
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		code = append(code, seg.Value(source)...)
	}

	w.WriteString(`<details class="doc-example"><summary>Example</summary>`)

	// Pick lexer: use the named language when one is specified and recognised;
	// fall back to the plain-text lexer for unspecified or unknown languages so
	// that HTML entities are escaped as contiguous tokens.
	var lexer chroma.Lexer
	if len(language) > 0 {
		lexer = lexers.Get(string(language))
	}
	if lexer == nil {
		lexer = defaultLexer
	}

	iter, err := lexer.Tokenise(nil, string(code))
	if err != nil || r.formatter.Format(w, r.style, iter) != nil {
		// Fallback: escape and write as plain code.
		w.WriteString(`<pre><code>`)
		w.Write(util.EscapeHTML(code))
		w.WriteString(`</code></pre>`)
	}

	w.WriteString(`</details>`)

	return ast.WalkSkipChildren, nil
}
