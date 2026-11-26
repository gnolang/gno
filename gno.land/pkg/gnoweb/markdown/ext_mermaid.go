package markdown

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown/mermaid"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// KindMermaid is the node kind of a ```mermaid fence drawn as a diagram.
var KindMermaid = ast.NewNodeKind("Mermaid")

// mermaidNode holds a ```mermaid fence, its only child, and the source the
// fence carries. The renderer draws the source as SVG, or shows the fence,
// through whatever renders code blocks, when it is not drawn.
type mermaidNode struct {
	ast.BaseBlock
	src    string
	budget *mermaidBudget
}

func (n *mermaidNode) Kind() ast.NodeKind { return KindMermaid }

func (n *mermaidNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// Per-document bounds: a document tries to draw at most maxMermaidTries
// fences, which with mermaid.MaxSourceLen bounds the drawing work of a
// render, and draws at most maxMermaidBytes of SVG. Fences past a bound show
// as code.
const (
	maxMermaidTries = 20
	maxMermaidBytes = 2 << 20
)

// mermaidBudget is what a document has drawn so far.
type mermaidBudget struct {
	bytes int
}

// mermaidTransformer turns the ```mermaid fences into mermaid nodes, when the
// render context enables diagrams (GnoContext.Mermaid). Like a ```math
// fence, the info string must be exactly "mermaid", and a blank fence stays
// a code block.
type mermaidTransformer struct{}

func (mermaidTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	if !getMermaidFromContext(pc) {
		return
	}
	src := reader.Source()
	var fences []*ast.FencedCodeBlock
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if len(fences) == maxMermaidTries {
			return ast.WalkStop, nil
		}
		fence, ok := n.(*ast.FencedCodeBlock)
		if ok && entering && fence.Info != nil && string(fence.Info.Segment.Value(src)) == "mermaid" &&
			!util.IsBlank(fence.Lines().Value(src)) {
			fences = append(fences, fence)
		}
		return ast.WalkContinue, nil
	})
	budget := &mermaidBudget{}
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
	svg, err := mermaid.Render(n.src)
	if err != nil || n.budget.bytes+len(svg) > maxMermaidBytes {
		return ast.WalkContinue, nil // show the fence
	}
	n.budget.bytes += len(svg)
	w.WriteString(`<div class="gno-mermaid">`)
	w.Write(svg)
	w.WriteString("</div>\n")
	return ast.WalkSkipChildren, nil
}

type mermaidExtension struct{}

// ExtMermaid draws ```mermaid fences as SVG diagrams. It is a no-op unless the
// render context enables it.
var ExtMermaid = &mermaidExtension{}

func (*mermaidExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(mermaidTransformer{}, 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(mermaidRenderer{}, 500),
	))
}
