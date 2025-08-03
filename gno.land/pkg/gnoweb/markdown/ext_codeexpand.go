package markdown

import (
	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// ExtCodeExpand is the extension that transforms code blocks into expandable details/summary elements
var ExtCodeExpand = &expandableCodeExtension{}

type expandableCodeExtension struct{}

// Cache for Chroma components to avoid recreating them
var (
	chromaFormatter *chromahtml.Formatter
	chromaStyle     *chroma.Style
	lexerCache      = make(map[string]chroma.Lexer)
)

func init() {
	// Initialize Chroma components
	chromaFormatter = chromahtml.New(
		chromahtml.WithClasses(true),
		chromahtml.ClassPrefix("chroma-"),
		chromahtml.WithLineNumbers(true),
		chromahtml.WithLinkableLineNumbers(true, "L"),
	)
	chromaStyle = styles.Get("friendly")
}

// expandableCodeRenderer renders expandable code blocks
type expandableCodeRenderer struct {
	html.Config
}

func (r *expandableCodeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.renderCodeBlock)
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
}

func (r *expandableCodeRenderer) renderCodeBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		// Start the expandable wrapper
		w.WriteString(`<details class="doc-example">`)
		w.WriteString(`<summary>Example</summary>`)
		w.WriteString(`<div class="">`)

		// Extract code block content
		var language []byte
		var lines *text.Segments

		switch codeBlock := n.(type) {
		case *ast.FencedCodeBlock:
			language = codeBlock.Language(source)
			lines = codeBlock.Lines()
		case *ast.CodeBlock:
			lines = codeBlock.Lines()
		default:
			return ast.WalkContinue, nil
		}

		// Build the code content
		var codeContent string
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			codeContent += string(line.Value(source))
		}

		// Use Chroma to highlight the code with cached components
		var lexer chroma.Lexer
		lang := "go"
		if len(language) > 0 {
			lang = string(language)
		}

		// Get lexer from cache or create new one
		if cachedLexer, exists := lexerCache[lang]; exists {
			lexer = cachedLexer
		} else {
			lexer = lexers.Get(lang)
			if lexer == nil {
				lexer = lexers.Get("go")
			}
			lexerCache[lang] = lexer
		}

		// Highlight the code with Chroma using cached components
		iterator, err := lexer.Tokenise(nil, codeContent)
		if err != nil || chromaFormatter.Format(w, chromaStyle, iterator) != nil {
			w.WriteString(`<pre><code class="language-go">`)
			w.WriteString(codeContent)
			w.WriteString(`</code></pre>`)
		}

		w.WriteString(`</div>`)
		w.WriteString(`</details>`)

		return ast.WalkSkipChildren, nil
	}

	return ast.WalkContinue, nil
}

func (e *expandableCodeExtension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&expandableCodeRenderer{}, 100),
		),
	)
}
