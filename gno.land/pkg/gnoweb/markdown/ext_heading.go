package markdown

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type anchorMode int

const (
	anchorNone anchorMode = iota
	anchorWrap
	anchorSibling
)

type headingRenderer struct{}

var _ renderer.NodeRenderer = (*headingRenderer)(nil)

func (r *headingRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHeading, r.renderHeading)
}

// renderHeading emits a heading with a clickable anchor link targeting its
// auto-generated id. Two modes:
//   - anchorWrap (no interactive descendants): wraps the heading text in
//     <a class="heading-anchor" href="#id">, so clicking any of the text
//     updates window.location.hash.
//   - anchorSibling (heading contains a link/autolink): emits
//     <a class="heading-anchor" href="#id"> with a visually-hidden label
//     after the heading text. Wrapping would produce nested <a> (invalid
//     HTML). The label keeps the sibling anchor accessible to screen
//     readers when focused via keyboard.
//
// When the heading has no usable id, no anchor is emitted.
func (r *headingRenderer) renderHeading(
	w util.BufWriter, source []byte, node ast.Node, entering bool,
) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	id, mode := headingAnchor(n)
	levelCh := "0123456"[n.Level]

	if entering {
		_, _ = w.WriteString("<h")
		_ = w.WriteByte(levelCh)
		if n.Attributes() != nil {
			html.RenderAttributes(w, node, html.HeadingAttributeFilter)
		}
		_ = w.WriteByte('>')
		if mode == anchorWrap {
			writeAnchorOpen(w, id)
		}
		return ast.WalkContinue, nil
	}

	switch mode {
	case anchorWrap:
		_, _ = w.WriteString(`</a>`)
	case anchorSibling:
		writeAnchorOpen(w, id)
		_, _ = w.WriteString(`<span class="u-sr-only">Permalink to this section</span></a>`)
	}
	_, _ = w.WriteString("</h")
	_ = w.WriteByte(levelCh)
	_, _ = w.WriteString(">\n")
	return ast.WalkContinue, nil
}

func writeAnchorOpen(w util.BufWriter, id []byte) {
	_, _ = w.WriteString(`<a class="heading-anchor" href="#`)
	_, _ = w.Write(util.EscapeHTML(id))
	_, _ = w.WriteString(`">`)
}

// headingAnchor returns the heading's anchor id and which anchor mode to
// use. Wrapping is skipped when the heading contains a link / autolink
// descendant (nested <a> is invalid HTML).
func headingAnchor(n *ast.Heading) ([]byte, anchorMode) {
	id, ok := n.AttributeString("id")
	if !ok {
		return nil, anchorNone
	}
	idBytes, ok := id.([]byte)
	if !ok || len(idBytes) == 0 {
		return nil, anchorNone
	}
	if hasLinkDescendant(n) {
		return idBytes, anchorSibling
	}
	return idBytes, anchorWrap
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
