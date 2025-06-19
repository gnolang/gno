package markdown

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"golang.org/x/net/html"
)

var KindForm = ast.NewNodeKind("Form")

// We only need two tags in our AST: the block node (gno-form)
// and the input nodes (gno-input).
type FormTag int

const (
	FormTagOpen  FormTag = iota // For the opening node (<gno-form>)
	FormTagInput                // For an input node (<gno-input />)
)

const (
	defaultInputType   = "text"
	defaultPlaceholder = "Enter value"
)

var (
	ErrFormUnexpectedOrInvalidTag = errors.New("unexpected or invalid tag")
	ErrFormInputMissingName       = errors.New("gno-input must have a 'name' attribute")
	ErrFormNoEndingTag            = errors.New("no ending tag found")
	ErrFormInvalidInputType       = errors.New("invalid input type")
)

// Whitelist of allowed input types
var allowedInputTypes = map[string]bool{
	"text":     true,
	"number":   true,
	"email":    true,
	"tel":      true,
	"password": true,
}

// validateInputType checks if the input type is allowed
func validateInputType(inputType string) bool {
	if inputType == "" {
		return true // Empty type will use default
	}

	return allowedInputTypes[inputType]
}

type FormInput struct {
	Error       error
	Name        string
	Type        string
	Placeholder string
}

func (in FormInput) String() string {
	if in.Error != nil {
		return fmt.Sprintf("(err=%s)", in.Error)
	}

	return fmt.Sprintf("(name=%s) (type=%s) (placeholder=%s)", in.Name, in.Type, in.Placeholder)
}

type FormNode struct {
	ast.BaseBlock
	Inputs     []FormInput
	RenderPath string // Path to render after form submission
	RealmName  string
}

func (n *FormNode) Kind() ast.NodeKind { return KindForm }

// Dump displays the information level for the Form node
func (n *FormNode) Dump(source []byte, level int) {
	kv := map[string]string{
		"path": n.RenderPath,
		"name": n.RealmName,
	}

	for i, in := range n.Inputs {
		kv[fmt.Sprintf("input_%d", i)] = in.String()
	}

	ast.DumpHelper(n, source, level, kv, nil)
}

func (n *FormNode) NewInput() (input *FormInput) {
	n.Inputs = append(n.Inputs, FormInput{})
	return &n.Inputs[len(n.Inputs)-1]
}

func (n *FormNode) NewErrorInput(err error) (input *FormInput) {
	input = n.NewInput()
	input.Error = err
	return input
}

// parseFormTag parses a form tag and returns the tag information
func parseFormTag(line []byte) (tok html.Token, ok bool) {
	line = bytes.TrimSpace(line)
	if len(line) > 0 {
		toks, err := ParseHTMLTokens(bytes.NewReader(line))
		if err == nil && len(toks) == 1 {
			return toks[0], true
		}
	}

	return
}

// formParser starts a block as soon as we encounter "<gno-form>"
// and closes it as soon as we encounter "</gno-form>".
// In between, only <gno-input /> lines are processed.
type formParser struct{}

var _ parser.BlockParser = (*formParser)(nil)

// NewFormParser creates a new instance of formParser
func NewFormParser() *formParser {
	return &formParser{}
}

// Trigger detects the start of the block
func (p *formParser) Trigger() []byte {
	return []byte{'<'}
}

// Open starts a block only when the line is exactly "<gno-form>"
func (p *formParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	tok, valid := parseFormTag(line)
	if !valid {
		return nil, parser.NoChildren // skip
	}

	if tok.Data != "gno-form" || tok.Type != html.StartTagToken {
		return nil, parser.NoChildren // skip, not our tag
	}

	fn := FormNode{}
	fn.RenderPath, _ = ExtractAttr(tok.Attr, "path")
	if gnourl, ok := getUrlFromContext(pc); ok {
		// Use full path instead of just namespace
		fn.RealmName = gnourl.Path
	}

	return &fn, parser.Continue
}

// Continue processes lines until "</gno-form>" is found.
// When a line contains <gno-input />, it adds a child node.
func (p *formParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, seg := reader.PeekLine()
	if seg.IsEmpty() {
		return parser.Continue
	}

	if line = bytes.TrimSpace(line); len(line) == 0 {
		return parser.Continue // skip empty line
	}

	formNode := node.(*FormNode)

	tok, valid := parseFormTag(line)
	if !valid {
		formNode.NewErrorInput(ErrFormUnexpectedOrInvalidTag)
		return parser.Continue
	}

	if tok.Data == "gno-form" {
		if tok.Type == html.EndTagToken {
			reader.AdvanceLine()
			return parser.Close // done
		}

		formNode.NewErrorInput(ErrFormUnexpectedOrInvalidTag)
		return parser.Continue
	}

	if tok.Data != "gno-input" {
		formNode.NewErrorInput(ErrFormUnexpectedOrInvalidTag)
		return parser.Continue
	}

	formInput := formNode.NewInput()
	if tok.Type != html.SelfClosingTagToken {
		formNode.NewErrorInput(ErrFormUnexpectedOrInvalidTag) // XXX: use better error
		return parser.Continue
	}

	for _, attr := range tok.Attr {
		switch attr.Key {
		case "name":
			formInput.Name = strings.TrimSpace(attr.Val)
		case "placeholder":
			formInput.Placeholder = strings.TrimSpace(attr.Val)
		case "type":
			formInput.Type = strings.TrimSpace(attr.Val)
		}
	}

	if formInput.Placeholder == "" {
		formInput.Placeholder = defaultPlaceholder
	}

	if formInput.Type == "" {
		formInput.Type = defaultInputType
	} else if !validateInputType(formInput.Type) {
		formInput.Error = ErrFormInvalidInputType
		return parser.Continue
	}

	if formInput.Name == "" {
		formInput.Error = ErrFormInputMissingName
		return parser.Continue
	}

	// Any other line (text, etc.) should stop the block
	return parser.Continue
}

// Close closes the block
func (p *formParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}

func (p *formParser) CanInterruptParagraph() bool { return true }
func (p *formParser) CanAcceptIndentedLine() bool { return true }

// formRenderer renders the Form node.
// When entering the Form node, it displays the opening <form> tag
// and when exiting (after rendering the child inputs),
// it displays the submit button and </form>.
type formRenderer struct{}

// NewFormRenderer creates a new instance of formRenderer
func NewFormRenderer() *formRenderer {
	return &formRenderer{}
}

// RegisterFuncs registers the render function for the Form node
func (r *formRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindForm, r.render)
}

// render renders the Form node
func (r *formRenderer) render(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n, ok := node.(*FormNode)
	if !ok {
		return ast.WalkContinue, nil
	}

	// Form action must include the full path
	formAction := n.RealmName // start with /r/docs/markdown
	if n.RenderPath != "" {
		formAction += ":" + strings.TrimPrefix(n.RenderPath, "/")
	}

	// Render form opening and header
	fmt.Fprintf(w, `<form class="gno-form" method="post" action="%s">`+"\n", HTMLEscapeString(formAction))
	fmt.Fprintln(w, `<div class="gno-form_header">`)
	fmt.Fprintf(w, `<span><span class="font-bold">%s</span>Form</span>`+"\n", HTMLEscapeString(n.RealmName))
	fmt.Fprintf(w, `<span class="tooltip" data-tooltip="Processed securely by %s"><svg class="w-4 h-4"><use href="#ico-info"></use></svg></span>`+"\n", HTMLEscapeString(n.RealmName))
	fmt.Fprintln(w, `</div>`)

	for _, in := range n.Inputs {
		if in.Error != nil {
			fmt.Fprintf(w, "<!-- Error: %s -->\n", HTMLEscapeString(in.Error.Error()))
			continue
		}

		// Render an input
		fmt.Fprintf(w, `<div class="gno-form_input"><label for="%s"> %s </label>`+"\n",
			HTMLEscapeString(in.Name),
			HTMLEscapeString(in.Placeholder))
		fmt.Fprintf(w, `<input type="%s" id="%s" name="%s" placeholder="%s" />`+"\n",
			HTMLEscapeString(in.Type),
			HTMLEscapeString(in.Name),
			HTMLEscapeString(in.Name),
			HTMLEscapeString(in.Placeholder))
		fmt.Fprintln(w, "</div>")
	}

	// Display submit button only if there is at least one input
	if len(n.Inputs) > 0 {
		fmt.Fprintf(w, `<input type="submit" value="Submit to %s Realm" />`+"\n", HTMLEscapeString(n.RealmName))
	}

	fmt.Fprintln(w, "</form>")
	return ast.WalkContinue, nil
}

type formExtension struct{}

// NewFormExtension creates a new instance of formExtension
func NewFormExtension() *formExtension {
	return &formExtension{}
}

// Extend adds parsing and rendering options for the Form node
func (e *formExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(util.Prioritized(NewFormParser(), 500)),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(util.Prioritized(NewFormRenderer(), 500)),
	)
}

// Forms is the extension instance
var ExtForms = &formExtension{}
