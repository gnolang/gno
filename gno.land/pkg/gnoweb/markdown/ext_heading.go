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
	href []byte // pre-escaped, ready to emit between `href="#` and `"`.
}

func (n *headingAnchorNode) Kind() ast.NodeKind { return KindHeadingAnchor }

func (n *headingAnchorNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"href": string(n.href)}, nil)
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
		// Headings are block nodes — skip into inline subtrees, they
		// can't contain headings and would be a wasted descent.
		if n.Type() == ast.TypeInline {
			return ast.WalkSkipChildren, nil
		}
		if h, ok := n.(*ast.Heading); ok {
			wrapHeadingChildren(h)
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
}

// wrapHeadingChildren regroups the heading's inline children so that each
// contiguous run of non-link nodes is moved under a fresh headingAnchorNode.
// Inline link nodes act as run boundaries and stay where they are.
// Pre-escapes the heading id once so the renderer doesn't re-escape per run.
//
// Idempotent: an existing headingAnchorNode child is treated as already-wrapped
// and skipped (it is neither a run boundary nor wrapped again).
func wrapHeadingChildren(h *ast.Heading) {
	id, ok := h.AttributeString("id")
	if !ok {
		return
	}
	idBytes, ok := id.([]byte)
	if !ok || len(idBytes) == 0 {
		return
	}
	href := util.EscapeHTML(idBytes)

	var run *headingAnchorNode
	c := h.FirstChild()
	for c != nil {
		next := c.NextSibling()
		switch {
		case c.Kind() == KindHeadingAnchor:
			// Already wrapped (re-running the transformer is a no-op).
			run = nil
		case isLinkLike(c):
			run = nil
		default:
			if run == nil {
				run = &headingAnchorNode{href: href}
				h.InsertBefore(h, c, run)
			}
			h.RemoveChild(h, c)
			run.AppendChild(run, c)
		}
		c = next
	}
}

// isLinkLike reports whether the node renders as an <a> tag — i.e. would
// produce nested <a> if wrapped by the heading-anchor.
//
// MAINTAINERS: when adding a new inline extension that emits an <a> element
// (footnote refs, card links, etc.), add its NodeKind here. Otherwise the
// heading-anchor will silently wrap it and produce nested anchors.
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
	_, _ = w.Write(n.href)
	_, _ = w.WriteString(`">`)
	return ast.WalkContinue, nil
}

type headingExtension struct{}

// ExtHeading is the heading-anchor extension instance, kept consistent
// with the package's other ExtXxx singletons (ExtLinks, ExtAlerts, …).
var ExtHeading = &headingExtension{}

// priorityHeadingAnchor must run after PriorityLinkTransformer so that
// linkTransformer has already rewritten Link/AutoLink → GnoLink — isLinkLike
// classifies children by their final kind. The +499 leaves a wide gap for
// any other transformer that needs to slot in between the two.
const priorityHeadingAnchor = PriorityLinkTransformer + 499

func (e *headingExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&headingAnchorTransformer{}, priorityHeadingAnchor),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&headingAnchorRenderer{}, 1),
	))
}
