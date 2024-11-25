package markdown

import (
	"github.com/yuin/goldmark/renderer"
)

type Renderer struct{}

// NewRenderer initialize Renderer as renderer.NodeRenderer.
func NewRenderer() renderer.NodeRenderer {
	return &Renderer{}
}

// RegisterFuncs add AST objects to Renderer.
func (r *Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindColumn, columnRender)
}

// func (r *Renderer) column(w util.BufWriter, source []byte, node) {
// 	ast.Node, entering bool) (ast.WalkStatus, error) {
// 	fmt.Println(node.Kind())
// 	return ast.WalkContinue, nil
// }
