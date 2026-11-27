package markdown

import (
	"bytes"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown/mermaid"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// KindMermaid is the node kind of a diagram: a ```mermaid fence or a
// <gno-mermaid> block.
var KindMermaid = ast.NewNodeKind("Mermaid")

// mermaidNode holds a diagram source and, as its only child, the code block
// the renderer shows when it does not draw the source: the ```mermaid fence
// it came from, or the body of a <gno-mermaid> block.
type mermaidNode struct {
	ast.BaseBlock
	// src is "" for a <gno-mermaid> block not to draw: one with another
	// attribute or value, or not ended by its close tag.
	src string
	// width is the width attribute of a <gno-mermaid> block: "", "fit" or
	// "full".
	width string
	// bad marks a <gno-mermaid> opener with another attribute or value.
	bad bool
	// budget is nil for a source not to draw.
	budget *mermaidBudget
}

func (n *mermaidNode) Kind() ast.NodeKind { return KindMermaid }

func (n *mermaidNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// Per-document bounds: a document tries to draw at most maxMermaidTries
// diagrams, which with mermaid.MaxSourceLen bounds the drawing work of a
// render, and draws at most maxMermaidBytes of SVG. Diagrams past a bound
// show as code. Only rendered diagrams count, so a fence in a footnote never
// referenced takes no try.
const (
	maxMermaidTries = 20
	maxMermaidBytes = 2 << 20
)

// mermaidBudget is what a document has tried and drawn so far.
type mermaidBudget struct {
	tries, bytes int
}

// mermaidTransformer gives the <gno-mermaid> blocks their budget and turns
// the ```mermaid fences into mermaid nodes, when the render context enables
// diagrams (GnoContext.Mermaid); otherwise both show as code. Like a ```math
// fence, the info string must be exactly "mermaid", and a blank fence stays
// a code block.
type mermaidTransformer struct{}

func (mermaidTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	if !getMermaidFromContext(pc) {
		return
	}
	src := reader.Source()
	budget := &mermaidBudget{}
	var fences []*ast.FencedCodeBlock
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *mermaidNode: // a <gno-mermaid> block
			if n.src != "" {
				n.budget = budget
			}
		case *ast.FencedCodeBlock:
			if n.Info != nil && string(n.Info.Segment.Value(src)) == "mermaid" && !util.IsBlank(n.Lines().Value(src)) {
				fences = append(fences, n)
			}
		}
		return ast.WalkContinue, nil
	})
	for _, fence := range fences {
		node := &mermaidNode{src: string(fence.Lines().Value(src)), budget: budget}
		parent := fence.Parent()
		parent.ReplaceChild(parent, fence, node)
		node.AppendChild(node, fence)
	}
}

type mermaidRenderer struct{}

func (mermaidRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindMermaid, renderMermaid)
}

func renderMermaid(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*mermaidNode)
	b := n.budget
	if b == nil || b.tries == maxMermaidTries {
		return ast.WalkContinue, nil // show the code
	}
	b.tries++
	svg, err := mermaid.Render(n.src)
	if err != nil || b.bytes+len(svg) > maxMermaidBytes {
		return ast.WalkContinue, nil
	}
	b.bytes += len(svg)
	w.WriteString(`<div class="gno-mermaid`)
	if n.width != "" {
		w.WriteString(" gno-mermaid--")
		w.WriteString(n.width) // "fit" or "full"
	}
	// Assistive technology reads the drawing as one image, labelled by its
	// title or type; the source follows it, visually hidden, as the text a
	// reader had without diagrams. It is also what the Copy button of a
	// rendered file copies (data-copy-exclude drops the drawing).
	w.WriteString(`"><div data-copy-exclude>`)
	w.Write(svg)
	w.WriteString(`</div><pre class="u-sr-only">`)
	w.Write(util.EscapeHTML([]byte(n.src)))
	w.WriteString("</pre></div>\n")
	return ast.WalkSkipChildren, nil
}

var (
	mermaidOpenPrefix  = []byte("<gno-mermaid")
	mermaidClosePrefix = []byte("</gno-mermaid")
)

// parseMermaidOpenTag reports whether a trimmed line is a <gno-mermaid>
// opener, and returns its width. valid is false for another attribute, a
// width other than fit or full, or a repeated one: the block still takes its
// body, which shows as code, rather than drawing what it was not asked to.
func parseMermaidOpenTag(line []byte) (tag, valid bool, width string) {
	valid = true
	size, selfClosing := scanGnoTag(line, mermaidOpenPrefix, len(line), func(key, val []byte) {
		if v := string(val); bytes.EqualFold(key, []byte("width")) && width == "" && (v == "fit" || v == "full") {
			width = v
		} else {
			valid = false
		}
	})
	if size == 0 || size < len(line) || selfClosing {
		return false, false, ""
	}
	return true, valid, width
}

// isMermaidClose reports whether a trimmed line is a </gno-mermaid> tag,
// read like the other gno-* close tags.
func isMermaidClose(line []byte) bool {
	attrs := 0
	size, selfClosing := scanGnoTag(line, mermaidClosePrefix, len(line), func(_, _ []byte) { attrs++ })
	return size > 0 && size == len(line) && !selfClosing && attrs == 0
}

// mermaidParser reads a <gno-mermaid> block. Like a fence, its body is
// collected verbatim, opaque to the other block parsers, up to a
// </gno-mermaid> line; a block its container or the document ends first is
// not closed, and shows as code.
type mermaidParser struct{}

func (mermaidParser) Trigger() []byte { return []byte{'<'} }

func (mermaidParser) Open(_ ast.Node, reader text.Reader, _ parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	tag, valid, width := parseMermaidOpenTag(trimTagLine(line))
	if !tag {
		return nil, parser.NoChildren
	}
	reader.AdvanceToEOL()
	n := &mermaidNode{width: width, bad: !valid}
	n.AppendChild(n, ast.NewCodeBlock())
	return n, parser.NoChildren
}

func (mermaidParser) Continue(node ast.Node, reader text.Reader, _ parser.Context) parser.State {
	n := node.(*mermaidNode)
	line, seg := reader.PeekLine()
	if len(line) == 0 { // the end of the document
		return parser.Close
	}
	reader.AdvanceToEOL()
	if isMermaidClose(trimTagLine(line)) {
		if !n.bad {
			n.src = string(n.FirstChild().Lines().Value(reader.Source()))
		}
		return parser.Close
	}
	n.FirstChild().Lines().Append(seg)
	return parser.Continue | parser.NoChildren
}

func (mermaidParser) Close(ast.Node, text.Reader, parser.Context) {}

// CanInterruptParagraph: true, like a fence, gno-details and gno-frame.
func (mermaidParser) CanInterruptParagraph() bool { return true }

// CanAcceptIndentedLine: false; 4+ spaces is an indented code block.
func (mermaidParser) CanAcceptIndentedLine() bool { return false }

type mermaidExtension struct{}

// ExtMermaid draws ```mermaid fences and <gno-mermaid> blocks as SVG
// diagrams. Unless the render context enables it, both show as code.
var ExtMermaid = &mermaidExtension{}

func (*mermaidExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(util.Prioritized(mermaidParser{}, 500)),
		parser.WithASTTransformers(util.Prioritized(mermaidTransformer{}, 500)),
	)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(mermaidRenderer{}, 500),
	))
}
