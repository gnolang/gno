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

const (
	defaultInputType       = "text"
	defaultPlaceholder     = "Enter value"
	defaultTextareaRows    = 4
	defaultTextareaMinRows = 2
	defaultTextareaMaxRows = 10

	// HTML tag names
	tagGnoForm     = "gno-form"
	tagGnoInput    = "gno-input"
	tagGnoTextarea = "gno-textarea"
	tagGnoSelect   = "gno-select"
)

var (
	ErrFormUnexpectedOrInvalidTag = errors.New("unexpected or invalid tag")
	ErrFormInputMissingName       = errors.New(tagGnoInput + " must have a 'name' attribute")
	ErrFormInvalidInputType       = errors.New("invalid input type")
	ErrFormInputNameAlreadyUsed   = errors.New("input name already used")
)

// Whitelist of allowed input types
var allowedInputTypes = map[string]bool{
	"text":     true,
	"number":   true,
	"email":    true,
	"tel":      true,
	"password": true,
	"radio":    true,
	"checkbox": true,
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
	Value       string // For radio/checkbox values
	Checked     bool   // For radio/checkbox checked state
	Description string // New attribute for description span
}

type FormTextarea struct {
	Error       error
	Name        string
	Placeholder string
	Rows        int
	Description string // New attribute for description span
}

type FormSelect struct {
	Error       error
	Name        string
	Value       string // Value and display text for this option
	Selected    bool   // Whether this option is selected
	Description string // New attribute for description span
}

// FormElement interface for form elements
type FormElement interface {
	GetError() error
}

func (in *FormInput) GetError() error    { return in.Error }
func (ta *FormTextarea) GetError() error { return ta.Error }
func (sel *FormSelect) GetError() error  { return sel.Error }

func (in FormInput) String() string {
	if in.Error != nil {
		return fmt.Sprintf("(err=%s)", in.Error)
	}

	base := fmt.Sprintf("(name=%s) (type=%s) (placeholder=%s)",
		in.Name, in.Type, in.Placeholder)

	if in.Type == "radio" || in.Type == "checkbox" {
		base += fmt.Sprintf(" (value=%s) (checked=%t)", in.Value, in.Checked)
	}

	if in.Description != "" {
		base += fmt.Sprintf(" (description=%s)", in.Description)
	}

	return base
}

func (ta FormTextarea) String() string {
	if ta.Error != nil {
		return fmt.Sprintf("(err=%s)", ta.Error)
	}

	base := fmt.Sprintf("(name=%s) (placeholder=%s) (rows=%d)", ta.Name, ta.Placeholder, ta.Rows)

	if ta.Description != "" {
		base += fmt.Sprintf(" (description=%s)", ta.Description)
	}

	return base
}

func (sel FormSelect) String() string {
	if sel.Error != nil {
		return fmt.Sprintf("(err=%s)", sel.Error)
	}

	base := fmt.Sprintf("(name=%s) (value=%s) (selected=%t)",
		sel.Name, sel.Value, sel.Selected)

	if sel.Description != "" {
		base += fmt.Sprintf(" (description=%s)", sel.Description)
	}

	return base
}

type FormNode struct {
	ast.BaseBlock
	Error        error
	Elements     []FormElement
	ElementsName map[string]bool
	RenderPath   string // Path to render after form submission
	RealmName    string
}

func NewFormNode() *FormNode {
	return &FormNode{
		ElementsName: map[string]bool{},
	}
}

func (n *FormNode) Kind() ast.NodeKind { return KindForm }

// Dump displays the information level for the Form node
func (n *FormNode) Dump(source []byte, level int) {
	kv := map[string]string{
		"path": n.RenderPath,
		"name": n.RealmName,
	}

	for i, element := range n.Elements {
		switch e := element.(type) {
		case *FormInput:
			kv[fmt.Sprintf("input_%d", i)] = e.String()
		case *FormTextarea:
			kv[fmt.Sprintf("textarea_%d", i)] = e.String()
		case *FormSelect:
			kv[fmt.Sprintf("select_%d", i)] = e.String()
		}
	}

	ast.DumpHelper(n, source, level, kv, nil)
}

func (n *FormNode) NewInput() (input *FormInput) {
	input = &FormInput{}
	n.Elements = append(n.Elements, input)
	return input
}

func (n *FormNode) NewErrorInput(err error) (input *FormInput) {
	input = n.NewInput()
	input.Error = err
	return input
}

func (n *FormNode) NewTextarea() (textarea *FormTextarea) {
	textarea = &FormTextarea{}
	n.Elements = append(n.Elements, textarea)
	return textarea
}

func (n *FormNode) NewErrorTextarea(err error) (textarea *FormTextarea) {
	textarea = n.NewTextarea()
	textarea.Error = err
	return textarea
}

func (n *FormNode) NewSelect() (sel *FormSelect) {
	sel = &FormSelect{}
	n.Elements = append(n.Elements, sel)
	return sel
}

func (n *FormNode) NewErrorSelect(err error) (sel *FormSelect) {
	sel = n.NewSelect()
	sel.Error = err
	return sel
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
	if !valid || tok.Data != tagGnoForm {
		return nil, parser.NoChildren
	}

	fn := NewFormNode()

	if tok.Type != html.StartTagToken {
		fn.Error = ErrFormUnexpectedOrInvalidTag
		return fn, parser.NoChildren // skip, not our tag
	}

	fn.RenderPath, _ = ExtractAttr(tok.Attr, "path")
	if gnourl, ok := getUrlFromContext(pc); ok {
		fn.RealmName = gnourl.Path // Use full path instead of just namespace
	}

	return fn, parser.Continue
}

// Continue processes lines until "</gno-form>" is found.
// When a line contains <gno-input />, it adds a child node.
func (p *formParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, _ := reader.PeekLine()
	if line = bytes.TrimSpace(line); len(line) == 0 {
		return parser.Continue // skip empty line
	}

	formNode := node.(*FormNode)

	tok, valid := parseFormTag(line)
	if !valid {
		formNode.NewErrorInput(ErrFormUnexpectedOrInvalidTag)
		return parser.Continue
	}

	if tok.Data == tagGnoForm {
		if tok.Type == html.EndTagToken {
			reader.AdvanceLine()
			return parser.Close // done
		}

		formNode.NewErrorInput(ErrFormUnexpectedOrInvalidTag)
		return parser.Continue
	}

	if tok.Data != tagGnoInput && tok.Data != tagGnoTextarea && tok.Data != tagGnoSelect {
		formNode.NewErrorInput(ErrFormUnexpectedOrInvalidTag)
		return parser.Continue
	}

	switch tok.Data {
	case tagGnoInput:
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
			case "description":
				formInput.Description = strings.TrimSpace(attr.Val)
			case "type":
				formInput.Type = strings.TrimSpace(attr.Val)
			case "value":
				// Value is required for radio and checkbox, optional for other types
				formInput.Value = strings.TrimSpace(attr.Val)
			case "checked":
				// Checked is only valid for radio and checkbox
				if formInput.Type != "" && formInput.Type != "radio" && formInput.Type != "checkbox" {
					formInput.Error = fmt.Errorf("'checked' attribute is only valid for radio and checkbox inputs, not for type '%s'", formInput.Type)
					return parser.Continue
				}
				formInput.Checked = strings.TrimSpace(attr.Val) == "true"
			}
		}

		if formInput.Name == "" {
			formInput.Error = ErrFormInputMissingName
			return parser.Continue
		}

		// For radio and checkbox, allow same name (needed for groups)
		// For other types, ensure unique names
		if formInput.Type != "radio" && formInput.Type != "checkbox" {
			if formNode.ElementsName[formInput.Name] {
				formInput.Error = fmt.Errorf("%q: %w", formInput.Name, ErrFormInputNameAlreadyUsed)
				return parser.Continue
			}
			formNode.ElementsName[formInput.Name] = true
		}

		// Set default placeholder only for non-radio/checkbox inputs
		if formInput.Placeholder == "" && formInput.Type != "radio" && formInput.Type != "checkbox" {
			formInput.Placeholder = defaultPlaceholder
		}

		if formInput.Type == "" {
			formInput.Type = defaultInputType
		} else if !validateInputType(formInput.Type) {
			formInput.Error = ErrFormInvalidInputType
			return parser.Continue
		}

		// Check if value is required for radio and checkbox
		if (formInput.Type == "radio" || formInput.Type == "checkbox") && formInput.Value == "" {
			formInput.Error = fmt.Errorf("'value' attribute is required for %s inputs", formInput.Type)
			return parser.Continue
		}

	case tagGnoTextarea:
		formTextarea := formNode.NewTextarea()
		if tok.Type != html.SelfClosingTagToken {
			formNode.NewErrorTextarea(ErrFormUnexpectedOrInvalidTag)
			return parser.Continue
		}

		for _, attr := range tok.Attr {
			switch attr.Key {
			case "name":
				formTextarea.Name = strings.TrimSpace(attr.Val)
			case "placeholder":
				formTextarea.Placeholder = strings.TrimSpace(attr.Val)
			case "rows":
				if _, err := fmt.Sscanf(attr.Val, "%d", &formTextarea.Rows); err != nil {
					formTextarea.Rows = defaultTextareaRows // default rows for textarea
				} else if formTextarea.Rows < defaultTextareaMinRows {
					formTextarea.Rows = defaultTextareaMinRows // min rows for textarea
				} else if formTextarea.Rows > defaultTextareaMaxRows {
					formTextarea.Rows = defaultTextareaMaxRows // max rows for textarea
				}
			case "description":
				formTextarea.Description = strings.TrimSpace(attr.Val)
			}
		}

		if formTextarea.Name == "" {
			formTextarea.Error = ErrFormInputMissingName
			return parser.Continue
		}

		if formNode.ElementsName[formTextarea.Name] {
			formTextarea.Error = fmt.Errorf("%q: %w", formTextarea.Name, ErrFormInputNameAlreadyUsed)
			return parser.Continue
		}
		formNode.ElementsName[formTextarea.Name] = true

		if formTextarea.Placeholder == "" {
			formTextarea.Placeholder = defaultPlaceholder
		}

		// Set default rows if not specified
		if formTextarea.Rows == 0 {
			formTextarea.Rows = defaultTextareaRows
		}

	case tagGnoSelect:
		formSelect := formNode.NewSelect()
		if tok.Type != html.SelfClosingTagToken {
			formNode.NewErrorSelect(ErrFormUnexpectedOrInvalidTag)
			return parser.Continue
		}

		for _, attr := range tok.Attr {
			switch attr.Key {
			case "name":
				formSelect.Name = strings.TrimSpace(attr.Val)
			case "value":
				formSelect.Value = strings.TrimSpace(attr.Val)
			case "description":
				formSelect.Description = strings.TrimSpace(attr.Val)
			case "selected":
				formSelect.Selected = strings.TrimSpace(attr.Val) == "true"
			}
		}

		if formSelect.Name == "" {
			formSelect.Error = ErrFormInputMissingName
			return parser.Continue
		}

	default:
		formNode.NewErrorInput(ErrFormUnexpectedOrInvalidTag)
		return parser.Continue
	}

	// Continue with the next line
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

	if n.Error != nil {
		fmt.Fprintf(w, "<!-- Error: %s -->\n", HTMLEscapeString(n.Error.Error()))
		return ast.WalkContinue, nil
	}

	// Form action must include the full path
	formAction := n.RealmName // start with /r/docs/markdown
	if n.RenderPath != "" {
		formAction += ":" + strings.TrimPrefix(n.RenderPath, "/")
	}

	// Render form opening and header
	fmt.Fprintf(w, `<form class="gno-form" method="post" action="%s" autocomplete="off" spellcheck="false">`+"\n", HTMLEscapeString(formAction))
	fmt.Fprintln(w, `<div class="gno-form_header">`)
	fmt.Fprintf(w, `<span><span class="font-bold">%s</span> Form</span>`+"\n", HTMLEscapeString(n.RealmName))
	fmt.Fprintf(w, `<span class="tooltip" data-tooltip="Processed securely by %s"><svg class="w-3 h-3"><use href="#ico-info"></use></svg></span>`+"\n", HTMLEscapeString(n.RealmName))
	fmt.Fprintln(w, `</div>`)

	// Render all form elements in order of appearance
	lastDescID := "" // Track the last description ID for aria-labelledby fallback

	for i, element := range n.Elements {
		if element.GetError() != nil {
			fmt.Fprintf(w, "<!-- Error: %s -->\n", HTMLEscapeString(element.GetError().Error()))
			continue
		}

		switch e := element.(type) {
		case *FormInput:
			// Show description span if available
			if e.Description != "" {
				descID := fmt.Sprintf("desc_%s_%d", e.Name, i)
				fmt.Fprintf(w, `<div id="%s" class="gno-form_description">%s</div>`+"\n",
					HTMLEscapeString(descID), HTMLEscapeString(e.Description))
				lastDescID = descID // Update last description ID
			}

			// Render different input types
			switch e.Type {
			case "radio", "checkbox":
				// Generate unique ID for radio/checkbox using index
				uniqueID := fmt.Sprintf("%s_%d", e.Name, i)

				fmt.Fprintf(w, `<div class="gno-form_selectable">`+"\n")
				fmt.Fprintf(w, `<input type="%s" id="%s" name="%s" value="%s"`,
					HTMLEscapeString(e.Type),
					HTMLEscapeString(uniqueID),
					HTMLEscapeString(e.Name),
					HTMLEscapeString(e.Value))
				if lastDescID != "" {
					fmt.Fprintf(w, ` aria-labelledby="%s"`, HTMLEscapeString(lastDescID))
				}
				if e.Checked {
					fmt.Fprintf(w, ` checked`)
				}
				fmt.Fprintf(w, ` />`+"\n")

				// Build label text: value + placeholder (if available)
				labelText := e.Value
				if e.Placeholder != "" {
					labelText += " - " + e.Placeholder
				}

				fmt.Fprintf(w, `<label for="%s"> %s </label>`+"\n",
					HTMLEscapeString(uniqueID),
					HTMLEscapeString(labelText))
				fmt.Fprintln(w, "</div>")

			default:
				// Render standard input
				fmt.Fprintf(w, `<div class="gno-form_input"><label for="%s"> %s </label>`+"\n",
					HTMLEscapeString(e.Name),
					HTMLEscapeString(e.Placeholder))
				fmt.Fprintf(w, `<input type="%s" id="%s" name="%s" placeholder="%s" />`+"\n",
					HTMLEscapeString(e.Type),
					HTMLEscapeString(e.Name),
					HTMLEscapeString(e.Name),
					HTMLEscapeString(e.Placeholder))
				fmt.Fprintln(w, "</div>")
			}

		case *FormTextarea:
			// Show description span if available
			if e.Description != "" {
				descID := fmt.Sprintf("desc_%s_%d", e.Name, i)
				fmt.Fprintf(w, `<div id="%s" class="gno-form_description">%s</div>`+"\n",
					HTMLEscapeString(descID), HTMLEscapeString(e.Description))
				lastDescID = descID // Update last description ID
			}

			fmt.Fprintf(w, `<div class="gno-form_input"><label for="%s"> %s </label>`+"\n",
				HTMLEscapeString(e.Name),
				HTMLEscapeString(e.Placeholder))
			fmt.Fprintf(w, `<textarea id="%s" name="%s" placeholder="%s" rows="%d"></textarea>`+"\n",
				HTMLEscapeString(e.Name),
				HTMLEscapeString(e.Name),
				HTMLEscapeString(e.Placeholder),
				e.Rows)
			fmt.Fprintln(w, "</div>")

		case *FormSelect:
			// Check if we already rendered a select for this name
			selectRendered := false
			for j := 0; j < i; j++ {
				if prevElement, ok := n.Elements[j].(*FormSelect); ok && prevElement.Name == e.Name {
					selectRendered = true
					break
				}
			}

			// If this is the first select element with this name, render the select container
			if !selectRendered {
				// Show description span if available (only for the first element)
				if e.Description != "" {
					descID := fmt.Sprintf("desc_%s_%d", e.Name, i)
					fmt.Fprintf(w, `<div id="%s" class="gno-form_description">%s</div>`+"\n",
						HTMLEscapeString(descID), HTMLEscapeString(e.Description))
					lastDescID = descID // Update last description ID
				}

				// Start the select container
				// Format the name to be more readable (capitalize first letter, replace underscores with spaces)
				labelText := strings.Title(strings.ReplaceAll(e.Name, "_", " "))
				fmt.Fprintf(w, `<div class="gno-form_select"><label for="%s"> %s </label>`+"\n",
					HTMLEscapeString(e.Name),
					HTMLEscapeString(labelText))
				fmt.Fprintf(w, `<select id="%s" name="%s"`,
					HTMLEscapeString(e.Name),
					HTMLEscapeString(e.Name))
				if lastDescID != "" {
					fmt.Fprintf(w, ` aria-labelledby="%s"`, HTMLEscapeString(lastDescID))
				}
				fmt.Fprintf(w, `>`+"\n")

				// Add a default option with the label
				article := GetWordArticle(labelText)
				fmt.Fprintf(w, `<option value="" >Select %s %s</option>`+"\n", article, HTMLEscapeString(labelText))

				// Collect all options for this select name
				for k := i; k < len(n.Elements); k++ {
					if optionElement, ok := n.Elements[k].(*FormSelect); ok && optionElement.Name == e.Name {
						fmt.Fprintf(w, `<option value="%s"`,
							HTMLEscapeString(optionElement.Value))
						if optionElement.Selected {
							fmt.Fprintf(w, ` selected`)
						}
						fmt.Fprintf(w, `>%s</option>`+"\n",
							HTMLEscapeString(optionElement.Value))
					}
				}

				// Close the select
				fmt.Fprintf(w, `</select>`+"\n")

				// SVG icon for select
				fmt.Fprintf(w, `<svg class="w-4 h-4 absolute right-2 top-1/2 -translate-y-1/2 pointer-events-none"><use href="#ico-arrow"></use></svg>`+"\n")

				fmt.Fprintln(w, "</div>")
			}
		}
	}

	// Display submit button only if there is at least one input or textarea
	if len(n.Elements) > 0 {
		fmt.Fprintf(w, `<input type="submit" value="Submit to %s Realm" />`+"\n", HTMLEscapeString(n.RealmName))
	}

	fmt.Fprintln(w, "</form>")
	return ast.WalkContinue, nil
}

type formExtension struct{}

// Extend adds parsing and rendering options for the Form node
func (e *formExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(util.Prioritized(NewFormParser(), 500)),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(util.Prioritized(NewFormRenderer(), 500)),
	)
}

var ExtForms = &formExtension{}
