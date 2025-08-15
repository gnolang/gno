package markdown

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

//--- Alert Types and Constants

// Alert represents a block-level alert element in markdown
// It can contain a header and content, and supports different alert types
type Alert struct {
	ast.BaseBlock
}

// Dump prints the AST structure for debugging purposes
func (n *Alert) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// KindAlert is the node kind identifier for Alert nodes
var KindAlert = ast.NewNodeKind("Alert")

// Kind returns the node kind identifier
func (n *Alert) Kind() ast.NodeKind {
	return KindAlert
}

// NewAlert creates a new Alert node
func NewAlert() *Alert {
	return &Alert{}
}

// AlertHeader represents the header part of an alert
// It contains the alert type and title
type AlertHeader struct {
	ast.BaseBlock
}

// Dump prints the AST structure for debugging purposes
func (n *AlertHeader) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// KindAlertHeader is the node kind identifier for AlertHeader nodes
var KindAlertHeader = ast.NewNodeKind("AlertHeader")

// Kind returns the node kind identifier
func (n *AlertHeader) Kind() ast.NodeKind {
	return KindAlertHeader
}

// NewAlertHeader creates a new AlertHeader node
func NewAlertHeader() *AlertHeader {
	return &AlertHeader{}
}

//--- Alert Components

// alertParser implements the parser for Alert blocks
// It recognizes lines starting with '>' followed by alert syntax
type alertParser struct{}

var defaultAlertParser = &alertParser{}

// NewAlertParser creates a new alert parser
func NewAlertParser() parser.BlockParser {
	return defaultAlertParser
}

// Trigger returns the byte that triggers this parser
func (b *alertParser) Trigger() []byte {
	return []byte{'>'}
}

// regex matches alert syntax: [!(type)]-(title)
var regex = regexp.MustCompile(`^\[!(?P<kind>[\w]+)\](?P<closed>-{0,1})($|\s+(?P<title>.*))`)

// process checks if the current line matches alert syntax
// Returns true if it's a valid alert line and the number of bytes to advance
func (b *alertParser) process(reader text.Reader) (bool, int) {
	line, _ := reader.PeekLine()
	w, pos := util.IndentWidth(line, reader.LineOffset())
	if w > 3 || pos >= len(line) || line[pos] != '>' {
		return false, 0
	}

	advanceBy := 1

	if pos+advanceBy >= len(line) || line[pos+advanceBy] == '\n' {
		return true, advanceBy
	}
	if line[pos+advanceBy] == ' ' || line[pos+advanceBy] == '\t' {
		advanceBy++
	}

	if line[pos+advanceBy-1] == '\t' {
		reader.SetPadding(2)
	}

	return true, advanceBy
}

const (
	AlertTypeNote AlertType = iota
	AlertTypeTip
	AlertTypeCaution
	AlertTypeWarning
	AlertTypeSuccess
	AlertTypeInfo
)

type AlertType int

func parseAlertType(kind string) (AlertType, string) {
	switch strings.ToLower(kind) {
	case "tip":
		return AlertTypeTip, "tip"
	case "caution":
		return AlertTypeCaution, "caution"
	case "warning":
		return AlertTypeWarning, "warning"
	case "success":
		return AlertTypeSuccess, "success"
	case "note":
		return AlertTypeNote, "note"
	default:
		return AlertTypeInfo, "info"
	}
}

// Open creates a new Alert node when alert syntax is detected
func (b *alertParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	ok, advanceBy := b.process(reader)
	if !ok {
		return nil, parser.NoChildren
	}

	line, _ := reader.PeekLine()
	if len(line) <= advanceBy {
		return nil, parser.NoChildren
	}

	subline := line[advanceBy:]
	if !regex.Match(subline) {
		return nil, parser.NoChildren
	}

	match := regex.FindSubmatch(subline)
	kind := match[1]
	closed := match[2]

	// Parse and validate alert type
	alertType, kindStr := parseAlertType(string(kind))

	alert := NewAlert()
	alert.SetAttributeString("kind", []byte(kindStr))
	alert.SetAttributeString("alertType", alertType)
	alert.SetAttributeString("closed", len(closed) != 0)

	i := strings.Index(string(line), "]")
	reader.Advance(i)

	return alert, parser.HasChildren
}

// Continue processes subsequent lines of an alert block
func (b *alertParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	ok, advanceBy := b.process(reader)
	if !ok {
		return parser.Close
	}

	reader.Advance(advanceBy)
	return parser.Continue | parser.HasChildren
}

// Close is called when the alert block ends
func (b *alertParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	// nothing to do
}

// CanInterruptParagraph indicates if this parser can interrupt a paragraph
func (b *alertParser) CanInterruptParagraph() bool {
	return true
}

// CanAcceptIndentedLine indicates if this parser accepts indented lines
func (b *alertParser) CanAcceptIndentedLine() bool {
	return false
}

// AlertHTMLRenderer implements the HTML renderer for Alert nodes
type AlertHTMLRenderer struct {
	html.Config
}

// NewAlertHTMLRenderer creates a new alert HTML renderer
func NewAlertHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &AlertHTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs registers the render functions
func (r *AlertHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindAlert, r.renderAlert)
}

// renderAlert renders an Alert node to HTML
func (r *AlertHTMLRenderer) renderAlert(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	var alertType string
	if t, ok := node.AttributeString("kind"); ok {
		alertType = string(t.([]uint8))
	}

	open := " open"
	if t, ok := node.AttributeString("closed"); ok {
		if t.(bool) {
			open = ""
		}
	}

	start := fmt.Sprintf(`<details class="gno-alert gno-alert-%s"%s>
`, strings.ToLower(alertType), open)

	if entering {
		w.WriteString(start)
	} else {
		w.WriteString("</div>\n</details>\n")
	}
	return ast.WalkContinue, nil
}

//--- AlertHeader Components
// AlertHeader represents the header part of an alert
// It contains the alert type and title

// alertHeaderParser implements the parser for AlertHeader blocks
type alertHeaderParser struct{}

var defaultAlertHeaderParser = &alertHeaderParser{}

// NewAlertHeaderParser creates a new alert header parser
func NewAlertHeaderParser() parser.BlockParser {
	return defaultAlertHeaderParser
}

// Trigger returns the byte that triggers this parser
func (b *alertHeaderParser) Trigger() []byte {
	return []byte{']'}
}

// Open creates a new AlertHeader node
func (b *alertHeaderParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	if parent.ChildCount() != 0 || parent.Kind() != KindAlert {
		return nil, parser.NoChildren
	}

	reader.Advance(1)
	next := reader.Peek()
	if next == '-' {
		reader.Advance(1)
	}

	_, segment := reader.Position()
	line, _ := reader.PeekLine()

	w, _ := util.IndentWidth(line, reader.LineOffset())
	reader.Advance(w)

	_, segment = reader.Position()
	line, _ = reader.PeekLine()

	if len(line) > 0 && line[len(line)-1] == '\n' {
		segment.Stop = segment.Stop - 1
	}

	alert := NewAlertHeader()

	// Always set the kind from the parent
	if t, ok := parent.AttributeString("kind"); ok {
		kind := string(t.([]uint8))
		alertType, kindStr := parseAlertType(kind)
		alert.SetAttributeString("kind", kindStr)
		alert.SetAttributeString("alertType", alertType)
	}

	if segment.Len() != 0 {
		segments := text.Segments{}
		segments.Append(segment)

		paragraph := ast.NewTextBlock()
		paragraph.SetLines(&segments)

		alert.AppendChild(alert, paragraph)
		alert.SetAttributeString("hasTitle", true)
	}

	return alert, parser.NoChildren
}

// Continue processes subsequent lines of an alert header
func (b *alertHeaderParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	return parser.Close
}

// Close is called when the alert header ends
func (b *alertHeaderParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	// nothing to do
}

// CanInterruptParagraph indicates if this parser can interrupt a paragraph
func (b *alertHeaderParser) CanInterruptParagraph() bool {
	return false
}

// CanAcceptIndentedLine indicates if this parser accepts indented lines
func (b *alertHeaderParser) CanAcceptIndentedLine() bool {
	return true
}

// AlertHeaderHTMLRenderer implements the HTML renderer for AlertHeader nodes
type AlertHeaderHTMLRenderer struct {
	html.Config
}

// NewAlertHeaderHTMLRenderer creates a new alert header HTML renderer
func NewAlertHeaderHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &AlertHeaderHTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs registers the render functions
func (r *AlertHeaderHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindAlertHeader, r.renderAlertHeader)
}

// renderAlertHeader renders an AlertHeader node to HTML
func (r *AlertHeaderHTMLRenderer) renderAlertHeader(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<summary>\n")
		var kind string
		if t, ok := node.AttributeString("kind"); ok {
			kind = strings.ToLower(t.(string))
			fmt.Fprintf(w, `<svg><use href="#ico-%s"></use></svg>`, kind)
			// Only show the kind if there's no explicit title
			if hasTitle, ok := node.AttributeString("hasTitle"); !ok || !hasTitle.(bool) {
				w.WriteString(titleCase(kind))
			}
		}
	} else {
		w.WriteString(`<svg><use href="#ico-arrow"></use></svg>`)
		w.WriteString("\n</summary>\n<div>\n")
	}
	return ast.WalkContinue, nil
}

//--- Extension

// alertExtension implements the Goldmark extension for alerts
type alertExtension struct{}

// ExtAlerts is the global instance of the alert extension
var ExtAlerts = &alertExtension{}

// Extend adds the alert parsers and renderers to the Goldmark instance
func (e *alertExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(NewAlertParser(), 799),
			util.Prioritized(NewAlertHeaderParser(), 799),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewAlertHTMLRenderer(), 0),
			util.Prioritized(NewAlertHeaderHTMLRenderer(), 0),
		),
	)
}
