package markdown

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type headingRenderer struct {
	html.Config
}

var _ renderer.NodeRenderer = (*headingRenderer)(nil)

func newHeadingRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &headingRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

func (r *headingRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHeading, r.renderHeading)
}

func (r *headingRenderer) renderHeading(
	w util.BufWriter, source []byte, node ast.Node, entering bool,
) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	if entering {
		_, _ = w.WriteString("<h")
		_ = w.WriteByte("0123456"[n.Level])
		if n.Attributes() != nil {
			html.RenderAttributes(w, node, html.HeadingAttributeFilter)
		}
		_ = w.WriteByte('>')
		id, hasID := n.AttributeString("id")
		if hasID {
			if idBytes, ok := id.([]byte); ok && len(idBytes) > 0 {
				_, _ = w.WriteString(`<a class="heading-anchor" href="#`)
				_, _ = w.Write(util.EscapeHTML(idBytes))
				_, _ = w.WriteString(`" aria-label="Link to this section">`)
			}
		}
	} else {
		_, _ = w.WriteString(`</a>`)
		_, _ = w.WriteString("</h")
		_ = w.WriteByte("0123456"[n.Level])
		_, _ = w.WriteString(">\n")
	}
	return ast.WalkContinue, nil
}

type headingExtension struct{}

var extHeading = &headingExtension{}

func (e *headingExtension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(newHeadingRenderer(), 1),
	))
}
