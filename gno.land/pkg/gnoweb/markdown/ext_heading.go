package markdown

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type headingRenderer struct{}

var _ renderer.NodeRenderer = (*headingRenderer)(nil)

func (r *headingRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHeading, r.renderHeading)
}

// renderHeading emits a heading with a clickable anchor link targeting its
// auto-generated id. Two modes:
//   - wrap mode (no interactive descendants): wraps the heading text in
//     <a class="heading-anchor" href="#id">, so clicking any of the text
//     updates window.location.hash.
//   - sibling mode (heading contains a link/autolink): emits an empty
//     <a class="heading-anchor" href="#id" aria-hidden="true"></a> after the
//     heading text. Wrapping would produce nested <a> (invalid HTML).
//
// When the heading has no usable id, no anchor is emitted.
func (r *headingRenderer) renderHeading(
	w util.BufWriter, source []byte, node ast.Node, entering bool,
) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	idBytes, wrap := anchorTarget(n)

	if entering {
		_, _ = w.WriteString("<h")
		_ = w.WriteByte("0123456"[n.Level])
		if n.Attributes() != nil {
			html.RenderAttributes(w, node, html.HeadingAttributeFilter)
		}
		_ = w.WriteByte('>')
		if wrap {
			_, _ = w.WriteString(`<a class="heading-anchor" href="#`)
			_, _ = w.Write(util.EscapeHTML(idBytes))
			_, _ = w.WriteString(`">`)
		}
		return ast.WalkContinue, nil
	}

	switch {
	case wrap:
		_, _ = w.WriteString(`</a>`)
	case len(idBytes) > 0:
		_, _ = w.WriteString(`<a class="heading-anchor" href="#`)
		_, _ = w.Write(util.EscapeHTML(idBytes))
		_, _ = w.WriteString(`" aria-hidden="true"></a>`)
	}
	_, _ = w.WriteString("</h")
	_ = w.WriteByte("0123456"[n.Level])
	_, _ = w.WriteString(">\n")
	return ast.WalkContinue, nil
}

// anchorTarget returns the heading's anchor id and whether the heading text
// can safely be wrapped in an <a>. Wrapping is skipped when the heading
// contains link / autolink descendants (nested <a> is invalid HTML).
func anchorTarget(n *ast.Heading) ([]byte, bool) {
	id, ok := n.AttributeString("id")
	if !ok {
		return nil, false
	}
	idBytes, ok := id.([]byte)
	if !ok || len(idBytes) == 0 {
		return nil, false
	}
	return idBytes, !hasLinkDescendant(n)
}

func hasLinkDescendant(root ast.Node) bool {
	var found bool
	_ = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || n == root {
			return ast.WalkContinue, nil
		}
		switch n.Kind() {
		case ast.KindLink, ast.KindAutoLink, KindGnoLink:
			found = true
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	return found
}

type headingExtension struct{}

var extHeading = &headingExtension{}

// Extend registers the heading renderer at priority 1 so it overrides
// goldmark's default HTML heading renderer (priority 10).
func (e *headingExtension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&headingRenderer{}, 1),
	))
}
