package markdown

import (
	"bytes"
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

// process checks if the current line matches alert syntax. It returns whether
// a `>` marker is there and the offset, FROM THE START OF THE LINE, of the
// first byte after the marker — the offset Open slices the line at to test the
// `[!KIND]` regex. Counting from the line start rather than from the marker is
// what lets an indented (up to three columns, per CommonMark) or padded marker
// open an alert; `pos` is also exactly the right argument for reader.Advance,
// which spends reader padding before it touches real bytes.
//
// This MUST NOT mutate the reader. Open calls it before the `[!KIND]` regex
// decides whether the line is an alert at all, and a block parser that
// declines to open has to leave the reader exactly as it found it —
// otherwise the parser that handles the line next inherits the change. See
// consumeMarker for why a stray padding is not survivable.
func (b *alertParser) process(reader text.Reader) (bool, int) {
	line, _ := reader.PeekLine()
	w, pos := util.IndentWidth(line, reader.LineOffset())
	if w > 3 || pos >= len(line) || line[pos] != '>' {
		return false, 0
	}

	pos++

	if pos >= len(line) || line[pos] == '\n' {
		return true, pos
	}
	if line[pos] == ' ' || line[pos] == '\t' {
		pos++
	}

	return true, pos
}

// consumeMarker consumes the `>` marker at the current reader position, plus
// one following space or tab, and reports whether a marker was there. It
// mirrors goldmark's blockquoteParser.process byte for byte, deliberately:
// the two parsers share the `>` trigger and alternate on nested `>` lines, so
// they have to agree on how far a marker advances the reader.
//
// The tab arithmetic is the part that matters. A tab is consumed and the
// REMAINDER of the tab stop is handed back as reader padding. Setting a fixed
// padding without consuming the tab instead leaves the reader pinned: padding
// units absorb the blockquote parser's own Advance calls, so
// parser.openBlocks re-opens a blockquote on the same byte and loops,
// appending an ast.Blockquote every pass until memory runs out. `>>>>>` plus a
// tab was enough to do it.
func consumeMarker(reader text.Reader) bool {
	line, _ := reader.PeekLine()
	w, pos := util.IndentWidth(line, reader.LineOffset())
	if w > 3 || pos >= len(line) || line[pos] != '>' {
		return false
	}
	pos++
	if pos >= len(line) || line[pos] == '\n' {
		reader.Advance(pos)
		return true
	}
	reader.Advance(pos)
	if line[pos] == ' ' || line[pos] == '\t' {
		padding := 0
		if line[pos] == '\t' {
			padding = util.TabWidth(reader.LineOffset()) - 1
		}
		reader.AdvanceAndSetPadding(1, padding)
	}
	return true
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
	ok, markerEnd := b.process(reader)
	if !ok {
		return nil, parser.NoChildren
	}

	line, _ := reader.PeekLine()
	if len(line) <= markerEnd {
		return nil, parser.NoChildren
	}

	subline := line[markerEnd:]
	if !regex.Match(subline) {
		return nil, parser.NoChildren
	}

	// Cross-family nesting cap (shared with gno-foreign, gno-columns).
	// On refusal, return nil so the `>` line falls through to the
	// blockquote parser.
	if !Push(pc) {
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

	// Advance to the `]` that closes `[!KIND]` so alertHeaderParser (trigger
	// `]`) takes over. Offset it from markerEnd, inside the subline the regex
	// actually matched, rather than searching the whole line.
	reader.Advance(markerEnd + bytes.IndexByte(subline, ']'))

	return alert, parser.HasChildren
}

// Continue processes subsequent lines of an alert block
func (b *alertParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	if !consumeMarker(reader) {
		return parser.Close
	}

	return parser.Continue | parser.HasChildren
}

// Close is called when the alert block ends. Pop the cross-family
// depth counter, matching the Push from Open. Goldmark calls Close
// exactly once per successfully-opened block (whether terminated
// normally or at EOF), so this pairs 1:1 with Open's Push.
func (b *alertParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	Pop(pc)
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

	line, _ := reader.PeekLine()

	// Advance by BYTE position, not by indent WIDTH. A tab between `]` and
	// the title has width 1-4 but occupies one byte, so advancing by the
	// width eats title bytes (`> [!NOTE]\ttitle` rendered as `tle`).
	_, pos := util.IndentWidth(line, reader.LineOffset())
	reader.Advance(pos)

	_, segment := reader.Position()
	line, _ = reader.PeekLine()

	// Drop the line terminator. Trimming only `\n` leaves a CRLF source with a
	// lone `\r`, which counts as a title and suppresses the default kind label,
	// so `> [!NOTE]\r\n` renders an empty <summary>. Subtract the bytes trimmed
	// rather than recomputing Stop, so any reader padding stays accounted for.
	if trimmed := bytes.TrimRight(line, "\r\n"); len(trimmed) != len(line) {
		segment.Stop -= len(line) - len(trimmed)
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
