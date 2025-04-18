package markdown

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var KindGnoMark = ast.NewNodeKind("GnoMarkBlock")

var templateRegistry = map[string]func(string) string{
	"petrinet": petriNetRender,
}

// REVIEW: consider moving this to a config
type WebHost struct {
	Base string
	Tag  string
	Path string
}

func (h *WebHost) Cdn() string {
	return h.Base + h.Tag + h.Path
}

type gnoMarkBlock struct {
	ast.BaseBlock
	Content string
}

var _ ast.Node = (*gnoMarkBlock)(nil)

func (b *gnoMarkBlock) Kind() ast.NodeKind {
	return KindGnoMark
}

func (b *gnoMarkBlock) Dump(source []byte, level int) {
	m := map[string]string{
		"Content": b.Content,
	}
	ast.DumpHelper(b, source, level, m, nil)
}

type gnoMarkParser struct{}

var gnoMarkStartTag = []byte("<gno-mark>")
var gnoMarkEndTag = []byte("</gno-mark>")

func (p *gnoMarkParser) Open(parent ast.Node, reader text.Reader, _ parser.Context) (ast.Node, parser.State) {
	_ = parent // REVIEW: any benefit to using parent?
	line, _ := reader.PeekLine()
	if !bytes.HasPrefix(line, gnoMarkStartTag) {
		return nil, parser.NoChildren
	}
	reader.AdvanceLine()
	return &gnoMarkBlock{}, parser.NoChildren
}

func (p *gnoMarkParser) Continue(node ast.Node, reader text.Reader, _ parser.Context) parser.State {
	line, _ := reader.PeekLine()
	if line == nil || bytes.HasPrefix(line, gnoMarkEndTag) {
		return parser.Close
	}
	block := node.(*gnoMarkBlock)
	block.Content += string(line)
	return parser.Continue
}

func (p *gnoMarkParser) Close(_ ast.Node, reader text.Reader, _ parser.Context) {
	for {
		line, _ := reader.PeekLine()
		if line == nil || bytes.HasPrefix(line, gnoMarkEndTag) {
			reader.AdvanceLine()
			break
		}
		reader.AdvanceLine()
	}
}

func (p *gnoMarkParser) CanInterruptParagraph() bool {
	return true
}

func (p *gnoMarkParser) CanAcceptIndentedLine() bool {
	return false
}

func (p *gnoMarkParser) Trigger() []byte {
	return []byte{'<'}
}

// gnoMarkRenderer renders the gnoMark block as HTML.
type gnoMarkRenderer struct {
	client *gnoweb.HTMLWebClient
}

type GnoMarkData struct {
	GnoMark string          `json:"gnoMark"`
	RawData json.RawMessage `json:"-"`
}

func (g *GnoMarkData) UnmarshalJSON(data []byte) error {
	var temp map[string]json.RawMessage
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if rawGnoMark, ok := temp["gnoMark"]; ok {
		if err := json.Unmarshal(rawGnoMark, &g.GnoMark); err != nil {
			return err
		}
	}

	g.RawData = data
	return nil
}

func (r *gnoMarkRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindGnoMark, r.renderGnoMarkBlock)
}

func (r *gnoMarkRenderer) renderGnoMarkBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	_ = source // source with tags
	if !entering {
		return ast.WalkContinue, nil
	}

	b, ok := node.(*gnoMarkBlock)
	if !ok {
		return ast.WalkContinue, nil
	}

	jsonContent := strings.TrimSuffix(b.Content, "<gno-mark>")

	var gnoMarkData GnoMarkData
	if err := gnoMarkData.UnmarshalJSON([]byte(jsonContent)); err != nil {
		return ast.WalkStop, err
	}

	template, ok := templateRegistry[gnoMarkData.GnoMark]

	if !ok {
		return ast.WalkStop, nil
	}

	_, _ = w.WriteString(template(jsonContent))

	return ast.WalkContinue, nil
}

// GnoMarkExtension is the Goldmark extension adding block parsers and renderers
// for GnoMark blocks: <gno-mark>...</gno-mark>
type GnoMarkExtension struct {
	Client *gnoweb.HTMLWebClient
}

func (e *GnoMarkExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(&gnoMarkParser{}, 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&gnoMarkRenderer{client: e.Client}, 500),
	))
}
