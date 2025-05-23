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
)

// --- Node Types and Constants ---

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

type FormNode struct {
	ast.BaseBlock
	Tag              FormTag
	InputName        string
	InputType        string
	InputPlaceholder string
	RealmName        string
	Error            error
}

func (n *FormNode) Kind() ast.NodeKind { return KindForm }

// Dump displays the information level for the Form node
func (n *FormNode) Dump(source []byte, level int) {
	kv := map[string]string{"tag": fmt.Sprintf("%v", n.Tag)}
	if n.Tag == FormTagInput {
		kv["name"] = n.InputName
	}
	ast.DumpHelper(n, source, level, kv, nil)
}

func NewForm(tag FormTag) *FormNode {
	return &FormNode{Tag: tag}
}

// --- Line Parsing Utility Function ---
//
// We do a very simplified parsing: we rely on the complete trim
// of the line to detect exact tags and, for <gno-input>, we extract
// in a rudimentary way the "name" and "placeholder" attributes (between quotes).
func parseFormTag(line []byte) (tag FormTag, name, placeholder, inputType string, err error) {
	trimmed := bytes.TrimSpace(line)
	// Start of form block
	if !(bytes.HasSuffix(trimmed, []byte(">")) || bytes.HasSuffix(trimmed, []byte("/>"))) {
		return 0, "", "", "", ErrFormUnexpectedOrInvalidTag
	}
	if bytes.Equal(trimmed, []byte("<gno-form>")) {
		return FormTagOpen, "", "", "", nil
	}
	// Close form block
	if bytes.Equal(trimmed, []byte("</gno-form>")) {
		// We don't have a closing node in our AST,
		// the closing tag only serves to end the block.
		return FormTagOpen, "", "", "", ErrFormNoEndingTag
	}
	// Input detection
	if bytes.HasPrefix(trimmed, []byte("<gno-input")) {
		// Simplified attribute extraction
		name = extractAttr(trimmed, "name")
		placeholder = extractAttr(trimmed, "placeholder")
		inputType = extractAttr(trimmed, "type")

		if strings.TrimSpace(name) == "" {
			// If "name" is missing, it's an error.
			return FormTagInput, "", placeholder, inputType, ErrFormInputMissingName
		}

		// Validate input type
		if !validateInputType(inputType) {
			return FormTagInput, name, placeholder, defaultInputType, ErrFormInvalidInputType
		}

		if inputType == "" {
			inputType = defaultInputType
		}

		return FormTagInput, name, placeholder, inputType, nil
	}
	return 0, "", "", "", ErrFormUnexpectedOrInvalidTag
}

// extractAttr extracts the value of the first found attribute from a line
func extractAttr(line []byte, attr string) string {
	pattern := attr + "=\""
	if i := bytes.Index(line, []byte(pattern)); i != -1 {
		start := i + len(pattern)
		if end := bytes.IndexByte(line[start:], '"'); end != -1 {
			return string(line[start : start+end])
		}
	}
	return ""
}

// --- The Block Parser ---

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
	line, seg := reader.PeekLine()
	tag, _, _, _, err := parseFormTag(line)
	if err != nil || tag != FormTagOpen {
		return nil, parser.NoChildren
	}
	node := NewForm(FormTagOpen)
	// Consume the line "<gno-form>"
	reader.Advance(seg.Len())
	return node, parser.HasChildren
}

// Continue processes lines until "</gno-form>" is encountered.
// When a line contains <gno-input />, it adds a child node.
func (p *formParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, seg := reader.PeekLine()
	if seg.Len() == 0 {
		return parser.Close
	}
	trimmed := bytes.TrimSpace(line)
	// If the line is exactly "</gno-form>", consume it and close the block.
	if bytes.Equal(trimmed, []byte("</gno-form>")) {
		reader.Advance(seg.Len())
		return parser.Close
	}
	// If the line starts with "<gno-input", we add an input node.
	tag, name, placeholder, inputType, err := parseFormTag(line)
	if tag == FormTagInput {
		input := NewForm(FormTagInput)

		// Get realm name and placeholder
		input.InputName = name
		input.InputType = inputType
		if placeholder == "" {
			input.InputPlaceholder = defaultPlaceholder
		} else {
			input.InputPlaceholder = placeholder
		}

		// If an error occurred during parsing, we store it in the node.
		if err != nil {
			input.Error = err
		}

		node.AppendChild(node, input)
		reader.Advance(seg.Len())
		return parser.Continue | parser.HasChildren
	}
	// Any other line (text, etc.) should stop the block
	return parser.Close
}

// Close closes the block
func (p *formParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}
func (p *formParser) CanInterruptParagraph() bool                                { return true }
func (p *formParser) CanAcceptIndentedLine() bool                                { return true }

// --- The Renderer ---

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
	n := node.(*FormNode)

	// Get realmName if defined, otherwise empty string.
	realmName := ""
	if n.RealmName != "" {
		realmName = "r/" + n.RealmName
	}

	if n.Tag == FormTagOpen {
		if entering {
			// Render form opening and header
			fmt.Fprintln(w, `<form class="gno-form" method="post">`)
			fmt.Fprintln(w, `<div class="gno-form_header">`)
			fmt.Fprintf(w, `<span><span class="font-bold"> %s </span> Form</span>`, realmName)
			fmt.Fprintf(w, `<span class="tooltip" data-tooltip="Processed securely by %s."><svg class="w-4 h-4"><use href="#ico-info"></use></svg></span>`, realmName)
			fmt.Fprintln(w, `</div>`)
		} else {
			// Check if the form contains at least one input
			hasInput := false
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				if formChild, ok := child.(*FormNode); ok && formChild.Tag == FormTagInput {
					hasInput = true
					break
				}
			}
			// Display submit button only if there is at least one input
			if hasInput {
				fmt.Fprintf(w, `<input type="submit" value="Submit to %s Realm" />`, realmName)
			}
			fmt.Fprintln(w, `</form>`)
		}
	} else if n.Tag == FormTagInput && entering {
		// Render an input
		fmt.Fprintf(w, `<div class="gno-form_input"><label for="%s"> %s </label>`, n.InputName, n.InputPlaceholder)
		fmt.Fprintf(w, `<input type="%s" id="%s" name="%s" placeholder="%s" /></div>`, n.InputType, n.InputName, n.InputName, n.InputPlaceholder)

		if n.Error != nil {
			fmt.Fprintf(w, `<!-- Error: %s -->`, n.Error.Error())
		}
	}
	return ast.WalkContinue, nil
}

// --- The AST Transformer ---

type formASTTransformer struct{}

// formASTTransformer is an AST transformer for the Form node
func (a *formASTTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	if gnourl, ok := getUrlFromContext(pc); ok {
		realm := gnourl.Namespace()
		// Traverse the AST to update open form nodes
		ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if fn, ok := n.(*FormNode); ok && entering && fn.Tag == FormTagOpen {
				fn.RealmName = realm
			}
			return ast.WalkContinue, nil
		})
	}
}

// --- The Goldmark Extension ---

type formExtension struct{}

// NewFormExtension creates a new instance of formExtension
func NewFormExtension() *formExtension {
	return &formExtension{}
}

// Extend adds parsing and rendering options for the Form node
func (e *formExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(util.Prioritized(NewFormParser(), 500)),
		parser.WithASTTransformers(util.Prioritized(&formASTTransformer{}, 500)),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(util.Prioritized(NewFormRenderer(), 500)),
	)
}

// Forms is the extension instance
var ExtForms = &formExtension{}
