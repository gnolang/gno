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
	level := "0123456"[n.Level]
	if entering {
		_, _ = w.WriteString("<h")
		_ = w.WriteByte(level)
		if n.Attributes() != nil {
			html.RenderAttributes(w, n, html.HeadingAttributeFilter)
		}
		_ = w.WriteByte('>')
		return ast.WalkContinue, nil
	}
	id, _ := n.AttributeString("id")
	if idBytes, ok := id.([]byte); ok && len(idBytes) > 0 && n.HasChildren() {
		// Same icon and label as the ui/pkg_anchor template. The word joiner,
		// with the anchor's nowrap, keeps the icon on the heading's last word.
		_, _ = w.WriteString(`<a href="#`)
		_, _ = w.Write(util.EscapeHTML(idBytes))
		_, _ = w.WriteString(`" class="heading-anchor" aria-label="Permalink">&#8288;<svg class="c-icon" aria-hidden="true"><use href="#ico-link"></use></svg></a>`)
	}
	_, _ = w.WriteString("</h")
	_ = w.WriteByte(level)
	_, _ = w.WriteString(">\n")
	return ast.WalkContinue, nil
}

type headingExtension struct{}

// ExtHeading closes every heading with a permalink anchor to its id.
var ExtHeading = &headingExtension{}

func (e *headingExtension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&headingRenderer{}, 500),
	))
}
