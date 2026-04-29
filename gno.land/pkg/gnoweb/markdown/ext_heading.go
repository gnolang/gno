package markdown

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// KindHeadingAnchor identifies the synthetic inline node that wraps a
// contiguous run of non-link inline children inside a heading. The
// transformer creates these so that the wrapped text is clickable to
// update window.location.hash, without producing nested <a> when the
// heading already contains an inline link.
var KindHeadingAnchor = ast.NewNodeKind("HeadingAnchor")

type headingAnchorNode struct {
	ast.BaseInline
	id []byte
}

func (n *headingAnchorNode) Kind() ast.NodeKind { return KindHeadingAnchor }

func (n *headingAnchorNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"id": string(n.id)}, nil)
}

// headingAnchorTransformer wraps each contiguous run of non-link inline
// children of a heading in a headingAnchorNode, so the wrapped text is
// clickable while existing inline links keep their own destination. A
// heading containing only link descendants gets no anchor wrap (nested
// <a> would be invalid HTML, and the inline link already carries the
// click target).
type headingAnchorTransformer struct{}

func (t *headingAnchorTransformer) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		wrapHeadingChildren(h)
		return ast.WalkSkipChildren, nil
	})
}

func wrapHeadingChildren(h *ast.Heading) {
	id, ok := h.AttributeString("id")
	if !ok {
		return
	}
	idBytes, ok := id.([]byte)
	if !ok || len(idBytes) == 0 {
		return
	}

	var run *headingAnchorNode
	c := h.FirstChild()
	for c != nil {
		next := c.NextSibling()
		if isLinkLike(c) {
			run = nil
		} else {
			if run == nil {
				run = &headingAnchorNode{id: idBytes}
				h.InsertBefore(h, c, run)
			}
			h.RemoveChild(h, c)
			run.AppendChild(run, c)
		}
		c = next
	}
}

// isLinkLike reports whether the node renders as an <a> tag — i.e. would
// produce nested <a> if wrapped by the heading-anchor. Extend this list
// when new link-producing inline kinds are added.
func isLinkLike(n ast.Node) bool {
	switch n.Kind() {
	case ast.KindLink, ast.KindAutoLink, KindGnoLink:
		return true
	}
	return false
}

type headingAnchorRenderer struct{}

func (r *headingAnchorRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindHeadingAnchor, r.render)
}

func (r *headingAnchorRenderer) render(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		_, _ = w.WriteString(`</a>`)
		return ast.WalkContinue, nil
	}
	n := node.(*headingAnchorNode)
	_, _ = w.WriteString(`<a class="heading-anchor" href="#`)
	_, _ = w.Write(util.EscapeHTML(n.id))
	_, _ = w.WriteString(`">`)
	return ast.WalkContinue, nil
}

type headingExtension struct{}

var extHeading = &headingExtension{}

func (e *headingExtension) Extend(m goldmark.Markdown) {
	// Run last (priority 999) so we observe the final inline tree —
	// in particular linkTransformer (priority 500) rewrites Link nodes
	// to GnoLink before we classify children with isLinkLike.
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&headingAnchorTransformer{}, 999),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&headingAnchorRenderer{}, 1),
	))
}
