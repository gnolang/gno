package markdown

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// headingRenderer replaces goldmark's heading renderer to close every heading
// that has an id and some content with one permalink anchor. The anchor wraps
// no heading content, so links, footnote refs and raw HTML in the heading are
// left as they are and the heading text stays selectable.
type headingRenderer struct{}

func (r *headingRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHeading, r.render)
}

func (r *headingRenderer) render(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	if entering {
		_, _ = w.WriteString("<h")
		_ = w.WriteByte("0123456"[n.Level])
		if n.Attributes() != nil {
			html.RenderAttributes(w, n, html.HeadingAttributeFilter)
		}
		_ = w.WriteByte('>')
		return ast.WalkContinue, nil
	}
	if id, ok := n.AttributeString("id"); ok && n.HasChildren() {
		if idBytes, ok := id.([]byte); ok && len(idBytes) > 0 {
			// Same icon and label as the ui/pkg_anchor template.
			_, _ = w.WriteString(`<a href="#`)
			_, _ = w.Write(util.EscapeHTML(idBytes))
			_, _ = w.WriteString(`" class="heading-anchor" aria-label="Permalink"><svg class="c-icon" aria-hidden="true"><use href="#ico-link"></use></svg></a>`)
		}
	}
	_, _ = w.WriteString("</h")
	_ = w.WriteByte("0123456"[n.Level])
	_, _ = w.WriteString(">\n")
	return ast.WalkContinue, nil
}

type headingExtension struct{}

// ExtHeading is the heading-anchor extension instance, kept consistent
// with the package's other ExtXxx singletons (ExtLinks, ExtAlerts, …).
var ExtHeading = &headingExtension{}

func (e *headingExtension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&headingRenderer{}, 500),
	))
}
