package markdown

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"golang.org/x/net/html"
)

// Error messages for invalid form tags.
var (
	ErrFormUnexpectedOrInvalidTag = errors.New("unexpected or invalid tag")
	ErrFormInputMissingName       = errors.New("gno-input must have a 'name' attribute")
	ErrFormNestedForm             = errors.New("nested forms are not allowed")
	ErrFormDuplicateAttribute     = errors.New("duplicate attribute in gno-input tag")
)

// Constants for input validation
const (
	defaultPlaceholder = "Enter value"
)

// Define custom node kind.
var KindForm = ast.NewNodeKind("Form")

// FormTag represents the type of tag in a form block.
type FormTag int

const (
	FormTagUndefined FormTag = iota
	FormTagOpen
	FormTagClose
	FormTagInput
)

var formTagNames = map[FormTag]string{
	FormTagUndefined: "FormTagUndefined",
	FormTagOpen:      "FormTagOpen",
	FormTagClose:     "FormTagClose",
	FormTagInput:     "FormTagInput",
}

// FormNode represents a semantic tree for a "form".
type FormNode struct {
	ast.BaseBlock
	Index       int     // Index of the form associated with the node.
	Tag         FormTag // Current Form Tag for this node.
	Error       error   // If not nil, indicates that the node is invalid.
	Name        string  // Name attribute for input fields
	Placeholder string  // Placeholder text for input fields

	ctx *formContext
}

// formContext is used to keep track of form's state across parsing.
type formContext struct {
	PrevContext    *formContext
	IsOpen         bool      // Indicates if a block has been correctly opened.
	Index          int       // Index of the current form; 0 indicates no form.
	OpenTag        *FormNode // First opening tag for this context.
	HasValidInputs bool      // Indicates if the form has at least one valid input
	RealmName      string    // Realm name from context
}

// Dump implements Node.Dump for debug representation.
func (n *FormNode) Dump(source []byte, level int) {
	kv := map[string]string{
		"tag": formTagNames[n.Tag],
	}
	if n.Tag == FormTagInput {
		kv["index"] = strconv.Itoa(n.Index)
	}
	if err := n.Error; err != nil {
		kv["error"] = err.Error()
	}

	ast.DumpHelper(n, source, level, kv, nil)
}

// Kind implements Node.Kind.
func (*FormNode) Kind() ast.NodeKind {
	return KindForm
}

func (n *FormNode) String() string {
	return formTagNames[n.Tag]
}

// NewForm initializes a FormNode object.
func NewForm(ctx *formContext, tag FormTag) *FormNode {
	return &FormNode{ctx: ctx, Tag: tag}
}

var formContextKey = parser.NewContextKey()

// parseLineTag identifies the tag type based on the line content.
// It returns a FormTag and a slice of comments if applicable.
func parseFormLineTag(line []byte) (FormTag, string, string, error) {
	line = util.TrimRightSpace(util.TrimLeftSpace(line))

	// Parse the line into HTML tokens
	toks, err := ParseHTMLTokens(bytes.NewReader(line))
	if err != nil || len(toks) != 1 {
		return FormTagUndefined, "", "", nil // Return early if error or no tokens
	}

	var tag FormTag
	var name string
	var placeholder string

	// Determine tag type based on the first token
	switch tok := toks[0]; tok.Data {
	case "gno-form":
		switch tok.Type {
		case html.StartTagToken:
			tag = FormTagOpen
		case html.EndTagToken:
			tag = FormTagClose
		}
	case "gno-input":
		tag = FormTagInput
		// Check for duplicate attributes only for gno-input
		seenAttrs := make(map[string]bool)
		for _, attr := range tok.Attr {
			if seenAttrs[attr.Key] {
				return FormTagUndefined, "", "", ErrFormDuplicateAttribute
			}
			seenAttrs[attr.Key] = true
			if attr.Key == "name" {
				name = attr.Val
			} else if attr.Key == "placeholder" {
				placeholder = attr.Val
			}
		}
	}

	return tag, name, placeholder, nil
}

// formParser implements BlockParser.
var _ parser.BlockParser = (*formParser)(nil)

type formParser struct{}

// Trigger returns the trigger characters for the parser.
func (*formParser) Trigger() []byte {
	return []byte{'<'}
}

// Open creates a form node based on the line tag.
func (p *formParser) Open(doc ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	// Columns tag cannot be a child of another node.
	if doc.Parent() != nil {
		return nil, parser.NoChildren
	}

	line, _ := reader.PeekLine()
	tag, name, placeholder, err := parseFormLineTag(line)
	if err != nil {
		node := NewForm(nil, FormTagUndefined)
		node.Error = err
		return node, parser.NoChildren
	}
	if tag == FormTagUndefined {
		return nil, parser.NoChildren
	}

	// Get form context.
	cctx, ok := pc.Get(formContextKey).(*formContext)
	if !ok || !cctx.IsOpen {
		cctx = &formContext{PrevContext: cctx}
		pc.Set(formContextKey, cctx)
	}

	node := NewForm(cctx, tag)
	node.Name = name
	if placeholder == "" {
		node.Placeholder = defaultPlaceholder
	} else {
		node.Placeholder = placeholder
	}

	switch tag {
	case FormTagOpen:
		if cctx.IsOpen {
			node.Error = ErrFormUnexpectedOrInvalidTag
			return node, parser.NoChildren
		}

		cctx.OpenTag = node
		cctx.IsOpen = true
		cctx.HasValidInputs = false // Reset valid inputs state when opening a new form

	case FormTagClose:
		if !cctx.IsOpen {
			node.Error = ErrFormUnexpectedOrInvalidTag
			return node, parser.NoChildren
		}

		cctx.IsOpen = false

	case FormTagInput:
		if !cctx.IsOpen {
			node.Error = ErrFormUnexpectedOrInvalidTag
			return node, parser.NoChildren
		}
		if name == "" {
			node.Error = ErrFormInputMissingName
			return node, parser.NoChildren
		}
		cctx.HasValidInputs = true // Mark that we have at least one valid input
	}

	return node, parser.NoChildren
}

// Continue returns the parser state for continued parsing.
// Not needed in form context.
func (*formParser) Continue(n ast.Node, reader text.Reader, _ parser.Context) parser.State {
	return parser.Close
}

// Close finalizes the parsing of the node.
// Not needed in form context.
func (*formParser) Close(_ ast.Node, reader text.Reader, _ parser.Context) {}

// CanInterruptParagraph determines if the parser can interrupt paragraphs.
func (*formParser) CanInterruptParagraph() bool {
	return true
}

// CanAcceptIndentedLine checks if the parser can handle indented lines.
func (*formParser) CanAcceptIndentedLine() bool {
	return true
}

// formASTTransformer implements ASTTransformer.
type formASTTransformer struct{}

// Transform modifies the AST to handle unfinished open tags and add context.
func (a *formASTTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	// Retrieve the form context
	cctx, ok := pc.Get(formContextKey).(*formContext)
	if !ok {
		return
	}

	// Check for unclosed tags
	if cctx.IsOpen {
		lc := doc.LastChild()
		nodeForm := NewForm(cctx, FormTagClose)
		doc.InsertAfter(doc, lc, nodeForm)
	}

	// Add realm name to form context
	if gnourl, ok := getUrlFromContext(pc); ok {
		cctx.RealmName = gnourl.Namespace()
	}
}

// formRendererHTML implements NodeRenderer.
type formRendererHTML struct{}

// RegisterFuncs adds AST objects to the Renderer.
func (r *formRendererHTML) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindForm, r.formRenderHTML)
}

// formRenderHTML renders the form node.
func (r *formRendererHTML) formRenderHTML(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	cnode, ok := node.(*FormNode)
	if !ok {
		return ast.WalkContinue, nil
	}

	// Check for any error
	if err := cnode.Error; err != nil {
		if entering {
			switch {
			case errors.Is(err, ErrFormUnexpectedOrInvalidTag):
				fmt.Fprintf(w, "<!-- unexpected/invalid %q omitted -->\n", cnode.String())
			case errors.Is(err, ErrFormInputMissingName) || errors.Is(err, ErrFormDuplicateAttribute):
				fmt.Fprintf(w, "<!-- gno-input error: %s -->\n", err.Error())
			default:
				fmt.Fprintf(w, "<!-- gno-form error: %s -->\n", err.Error())
			}
		}
		return ast.WalkContinue, nil
	}

	realmName := "r/" + html.EscapeString(cnode.ctx.RealmName)

	// Render the node
	switch cnode.Tag {
	case FormTagOpen:
		if entering {
			fmt.Fprintln(w, `<form class="gno-form" method="post">`)
			fmt.Fprintln(w, `<div class="gno-form_header">`)
			fmt.Fprintf(w, `<span><span class="font-bold"> %s </span> Form</span>`, realmName)
			fmt.Fprintf(w, `<span class="tooltip" data-tooltip="This form is secure and processed on the current %s realm itself."><svg class="w-4 h-4"><use href="#ico-info"></use></svg></span>`, realmName)

			fmt.Fprintln(w, `</div>`)
		}
	case FormTagClose:
		if !entering {
			// Only show submit button if there are valid inputs
			if cnode.ctx.HasValidInputs {
				fmt.Fprintf(w, `<input type="submit" value="Submit to %s Realm" />`, realmName)
			}
			fmt.Fprintln(w, "</form>")
		}
	case FormTagInput:
		if entering {
			// Escape the placeholder and name
			placeholder := html.EscapeString(cnode.Placeholder)
			name := html.EscapeString(cnode.Name)

			fmt.Fprintf(w, `<div class="gno-form_input"><label for="%s"> %s </label>`, name, placeholder)
			fmt.Fprintf(w, `<input type="text" id="%s" name="%s" placeholder="%s" />`, name, name, placeholder)
			fmt.Fprintf(w, `</div>`)
		}
	default:
		panic("invalid form tag - should not happen")
	}

	return ast.WalkContinue, nil
}

type forms struct{}

// Forms instance for extending markdown with form functionality.
var Forms = &forms{}

func (e *forms) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(&formParser{}, 500),
		),
		parser.WithASTTransformers(
			util.Prioritized(&formASTTransformer{}, 500),
		),
	)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&formRendererHTML{}, 500),
	))
}
