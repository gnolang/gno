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
	RenderPath       string // Path to render after form submission
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

// FormTagInfo contains all the information parsed from a form tag
type FormTagInfo struct {
	Tag         FormTag
	Name        string
	Placeholder string
	InputType   string
	RenderPath  string
	Error       error
}

// parseFormTag parses a form tag and returns the tag information
func parseFormTag(line []byte) FormTagInfo {
	trimmed := bytes.TrimSpace(line)
	result := FormTagInfo{}

	// Start of form block
	if !(bytes.HasSuffix(trimmed, []byte(">")) || bytes.HasSuffix(trimmed, []byte("/>"))) {
		result.Error = ErrFormUnexpectedOrInvalidTag
		return result
	}

	if bytes.Equal(trimmed, []byte("<gno-form>")) {
		result.Tag = FormTagOpen
		return result
	}

	// Close form block
	if bytes.Equal(trimmed, []byte("</gno-form>")) {
		result.Tag = FormTagOpen
		result.Error = ErrFormNoEndingTag
		return result
	}

	// Input detection
	if bytes.HasPrefix(trimmed, []byte("<gno-input")) {
		result.Tag = FormTagInput
		result.Name = extractAttr(trimmed, "name")
		result.Placeholder = extractAttr(trimmed, "placeholder")
		result.InputType = extractAttr(trimmed, "type")

		// Always set default type, even if name is missing
		if result.InputType == "" {
			result.InputType = defaultInputType
		}

		if strings.TrimSpace(result.Name) == "" {
			// If "name" is missing, it's an error, but keep the type
			result.Error = ErrFormInputMissingName
			return result
		}

		// Validate input type
		if !validateInputType(result.InputType) {
			result.InputType = defaultInputType
			result.Error = ErrFormInvalidInputType
			return result
		}

		return result
	}

	// Extract path attribute for form tag
	if bytes.HasPrefix(trimmed, []byte("<gno-form")) {
		result.Tag = FormTagOpen
		result.RenderPath = extractAttr(trimmed, "path")
		return result
	}

	result.Error = ErrFormUnexpectedOrInvalidTag
	return result
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
	info := parseFormTag(line)
	if info.Error != nil || info.Tag != FormTagOpen {
		return nil, parser.NoChildren
	}
	
	node := NewForm(FormTagOpen)
	node.RenderPath = info.RenderPath
	reader.Advance(seg.Len()) // Consume the line "<gno-form>"
	return node, parser.HasChildren
}

// Continue processes lines until "</gno-form>" is found.
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
	info := parseFormTag(line)
	if info.Tag == FormTagInput {
		input := NewForm(FormTagInput)

		// Get realm name and placeholder
		input.InputName = info.Name
		input.InputType = info.InputType
		input.InputPlaceholder = info.Placeholder
		if info.Placeholder == "" {
			input.InputPlaceholder = defaultPlaceholder
		} 

		// If an error occurred during parsing, we store it in the node.
		if info.Error != nil {
			input.Error = info.Error
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

// --- Renderer ---

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

	if n.Tag == FormTagOpen {
		if entering {
			// Form action must include the full path
			formAction := n.RealmName // start with /r/docs/markdown
			if n.RenderPath != "" {
				formAction += ":" + strings.TrimPrefix(n.RenderPath, "/")
			}

			// Render form opening and header
			fmt.Fprintf(w, `<form class="gno-form" method="post" action="%s">`, formAction)
			fmt.Fprintln(w, `<div class="gno-form_header">`)
			fmt.Fprintf(w, `<span><span class="font-bold"> %s </span> Form</span>`, n.RealmName)
			fmt.Fprintf(w, `<span class="tooltip" data-tooltip="Processed securely by %s"><svg class="w-4 h-4"><use href="#ico-info"></use></svg></span>`, n.RealmName)
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
				fmt.Fprintf(w, `<input type="submit" value="Submit to %s Realm" />`, n.RealmName)
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
		// Use full path instead of just namespace
		realm := gnourl.Path
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
