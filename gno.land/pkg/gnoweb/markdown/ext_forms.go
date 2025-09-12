package markdown

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"golang.org/x/net/html"
)

// Form-specific constants
const (
	formTagName     = "gno-form"
	formInputTag    = "gno-input"
	formTextareaTag = "gno-textarea"
	formSelectTag   = "gno-select"

	formDefaultInputType    = "text"
	formDefaultPlaceholder  = "Enter value"
	formDefaultTextareaRows = 4
	formMinTextareaRows     = 2
	formMaxTextareaRows     = 10
)

// Form-specific errors
var (
	ErrFormInvalidTag        = errors.New("unexpected or invalid tag")
	ErrFormMissingName       = errors.New("missing 'name' attribute")
	ErrFormInvalidInputType  = errors.New("invalid input type")
	ErrFormDuplicateName     = errors.New("name already used")
	ErrFormInvalidAttribute  = errors.New("invalid attribute for input type")
	ErrFormMissingValue      = errors.New("missing 'value' attribute")
	ErrFormParameterNotFound = errors.New("parameter not found in function")
	ErrFormUnsupportedType   = errors.New("unsupported parameter type")
)

var (
	FormKind = ast.NewNodeKind("Form")

	formAllowedInputTypes = map[string]bool{
		"text":     true,
		"number":   true,
		"email":    true,
		"tel":      true,
		"password": true,
		"radio":    true,
		"checkbox": true,
	}
)

// FormElement represents any form element
type FormElement interface {
	GetName() string
	GetError() error
	String() string
}

// FormInput represents an input element
type FormInput struct {
	Name        string
	Type        string
	Placeholder string
	Value       string
	Checked     bool
	Description string
	Error       error
}

func (e FormInput) GetName() string { return e.Name }
func (e FormInput) GetError() error { return e.Error }
func (e FormInput) String() string {
	if e.Error != nil {
		return fmt.Sprintf("(err=%s)", e.Error)
	}
	s := fmt.Sprintf("(name=%s) (type=%s)", e.Name, e.Type)
	if e.Type != "radio" && e.Type != "checkbox" {
		s += fmt.Sprintf(" (placeholder=%s)", e.Placeholder)
	} else {
		s += fmt.Sprintf(" (value=%s) (checked=%t)", e.Value, e.Checked)
	}
	if e.Description != "" {
		s += fmt.Sprintf(" (description=%s)", e.Description)
	}
	return s
}

// FormTextarea represents a textarea element
type FormTextarea struct {
	Name        string
	Placeholder string
	Rows        int
	Description string
	Error       error
}

func (e FormTextarea) GetName() string { return e.Name }
func (e FormTextarea) GetError() error { return e.Error }
func (e FormTextarea) String() string {
	if e.Error != nil {
		return fmt.Sprintf("(err=%s)", e.Error)
	}
	s := fmt.Sprintf("(name=%s) (placeholder=%s) (rows=%d)", e.Name, e.Placeholder, e.Rows)
	if e.Description != "" {
		s += fmt.Sprintf(" (description=%s)", e.Description)
	}
	return s
}

// FormSelect represents a select option
type FormSelect struct {
	Name        string
	Value       string
	Selected    bool
	Description string
	Error       error
}

func (e FormSelect) GetName() string { return e.Name }
func (e FormSelect) GetError() error { return e.Error }
func (e FormSelect) String() string {
	if e.Error != nil {
		return fmt.Sprintf("(err=%s)", e.Error)
	}
	s := fmt.Sprintf("(name=%s) (value=%s) (selected=%t)", e.Name, e.Value, e.Selected)
	if e.Description != "" {
		s += fmt.Sprintf(" (description=%s)", e.Description)
	}
	return s
}

// FormNode represents a form in the AST
type FormNode struct {
	ast.BaseBlock
	Elements   []FormElement
	Exec       *vm.FunctionSignature
	RenderPath string
	RealmName  string
	Error      error
	usedNames  map[string]bool
}

func (n *FormNode) Kind() ast.NodeKind { return FormKind }

func (n *FormNode) Dump(source []byte, level int) {
	kv := map[string]string{
		"path": n.RenderPath,
		"name": n.RealmName,
	}
	for i, element := range n.Elements {
		kv[fmt.Sprintf("element_%d", i)] = element.String()
	}
	ast.DumpHelper(n, source, level, kv, nil)
}

func (n *FormNode) addElement(elem FormElement) {
	n.Elements = append(n.Elements, elem)
}

func (n *FormNode) validateName(name string, elemType string) error {
	if name == "" {
		return ErrFormMissingName
	}
	// Allow duplicate names for radio and checkbox inputs
	if elemType != "radio" && elemType != "checkbox" {
		if n.usedNames[name] {
			return fmt.Errorf("%q: %w", name, ErrFormDuplicateName)
		}
		n.usedNames[name] = true
	}
	return nil
}

// FormParser handles parsing of form blocks
type FormParser struct{}

func NewFormParser() *FormParser { return &FormParser{} }

func (p *FormParser) Trigger() []byte { return []byte{'<'} }

func (p *FormParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	tok, ok := parseFormTag(line)
	if !ok || tok.Data != formTagName || tok.Type != html.StartTagToken {
		return nil, parser.NoChildren
	}

	node := &FormNode{usedNames: make(map[string]bool)}

	// Extract attributes
	node.RenderPath, _ = ExtractAttr(tok.Attr, "path")
	if gnourl, ok := getUrlFromContext(pc); ok {
		node.RealmName = gnourl.Path
	}

	// Handle exec attribute
	if exec, ok := ExtractAttr(tok.Attr, "exec"); ok {
		if sigGetter, ok := getRealmFuncsGetterFromContext(pc); ok {
			sig, err := sigGetter(exec)
			if err != nil || sig == nil {
				node.Error = ErrFormInvalidTag
			} else {
				node.Exec = sig
			}
		}
	}

	return node, parser.Continue
}

func (p *FormParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, _ := reader.PeekLine()
	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return parser.Continue
	}

	formNode := node.(*FormNode)
	tok, ok := parseFormTag(line)
	if !ok {
		formNode.addElement(FormInput{Error: ErrFormInvalidTag})
		return parser.Continue
	}

	// Check for closing tag
	if tok.Data == formTagName {
		if tok.Type == html.EndTagToken {
			reader.AdvanceLine()
			return parser.Close
		}
		formNode.addElement(FormInput{Error: ErrFormInvalidTag})
		return parser.Continue
	}

	// Process form elements
	switch tok.Data {
	case formInputTag:
		p.parseInput(formNode, tok)
	case formTextareaTag:
		p.parseTextarea(formNode, tok)
	case formSelectTag:
		p.parseSelect(formNode, tok)
	default:
		formNode.addElement(FormInput{Error: ErrFormInvalidTag})
	}

	return parser.Continue
}

func (p *FormParser) parseInput(node *FormNode, tok html.Token) {
	if tok.Type != html.SelfClosingTagToken {
		node.addElement(FormInput{Error: ErrFormInvalidTag})
		return
	}

	input := FormInput{Type: formDefaultInputType}
	attrs := make(map[string]string)

	// Collect attributes
	for _, attr := range tok.Attr {
		attrs[attr.Key] = strings.TrimSpace(attr.Val)
	}

	// Process attributes
	input.Name = attrs["name"]
	if t := attrs["type"]; t != "" {
		input.Type = t
	}
	input.Placeholder = attrs["placeholder"]
	input.Description = attrs["description"]
	input.Value = attrs["value"]
	input.Checked = attrs["checked"] == "true"

	// Validate
	if err := node.validateName(input.Name, input.Type); err != nil {
		input.Error = err
		node.addElement(input)
		return
	}

	if !formAllowedInputTypes[input.Type] {
		input.Error = ErrFormInvalidInputType
		node.addElement(input)
		return
	}

	// Type-specific validation
	isSelectable := input.Type == "radio" || input.Type == "checkbox"

	if attrs["checked"] != "" && !isSelectable {
		input.Error = fmt.Errorf("'checked' attribute: %w for type '%s'", ErrFormInvalidAttribute, input.Type)
		node.addElement(input)
		return
	}

	if isSelectable && input.Value == "" {
		input.Error = fmt.Errorf("%w for %s input", ErrFormMissingValue, input.Type)
		node.addElement(input)
		return
	}

	// Set defaults
	if input.Placeholder == "" && !isSelectable {
		input.Placeholder = formDefaultPlaceholder
	}

	node.addElement(input)
}

func (p *FormParser) parseTextarea(node *FormNode, tok html.Token) {
	if tok.Type != html.SelfClosingTagToken {
		node.addElement(FormTextarea{Error: ErrFormInvalidTag})
		return
	}

	textarea := FormTextarea{
		Placeholder: formDefaultPlaceholder,
		Rows:        formDefaultTextareaRows,
	}

	// Process attributes
	for _, attr := range tok.Attr {
		switch attr.Key {
		case "name":
			textarea.Name = strings.TrimSpace(attr.Val)
		case "placeholder":
			if p := strings.TrimSpace(attr.Val); p != "" {
				textarea.Placeholder = p
			}
		case "rows":
			if _, err := fmt.Sscanf(attr.Val, "%d", &textarea.Rows); err != nil {
				textarea.Rows = formDefaultTextareaRows
			}
			// Clamp rows value
			if textarea.Rows < formMinTextareaRows {
				textarea.Rows = formMinTextareaRows
			} else if textarea.Rows > formMaxTextareaRows {
				textarea.Rows = formMaxTextareaRows
			}
		case "description":
			textarea.Description = strings.TrimSpace(attr.Val)
		}
	}

	// Validate
	if err := node.validateName(textarea.Name, "textarea"); err != nil {
		textarea.Error = err
	}

	node.addElement(textarea)
}

func (p *FormParser) parseSelect(node *FormNode, tok html.Token) {
	if tok.Type != html.SelfClosingTagToken {
		node.addElement(FormSelect{Error: ErrFormInvalidTag})
		return
	}

	sel := FormSelect{}

	for _, attr := range tok.Attr {
		switch attr.Key {
		case "name":
			sel.Name = strings.TrimSpace(attr.Val)
		case "value":
			sel.Value = strings.TrimSpace(attr.Val)
		case "selected":
			sel.Selected = strings.TrimSpace(attr.Val) == "true"
		case "description":
			sel.Description = strings.TrimSpace(attr.Val)
		}
	}

	if sel.Name == "" {
		sel.Error = ErrFormMissingName
	}

	node.addElement(sel)
}

func (p *FormParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}
func (p *FormParser) CanInterruptParagraph() bool                                { return true }
func (p *FormParser) CanAcceptIndentedLine() bool                                { return true }

// FormRenderer handles rendering of form nodes
type FormRenderer struct{}

func NewFormRenderer() *FormRenderer { return &FormRenderer{} }

func (r *FormRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(FormKind, r.render)
}

func (r *FormRenderer) render(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
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

	// Build form action
	action := n.RealmName
	if n.RenderPath != "" {
		action += ":" + strings.TrimPrefix(n.RenderPath, "/")
	}

	// Render form opening
	fmt.Fprintf(w, `<form class="gno-form" method="post" action="%s" autocomplete="off" spellcheck="false">`+"\n",
		HTMLEscapeString(action))
	fmt.Fprintf(w, `<div class="gno-form_header">
<span><span class="font-bold">%s</span> Form</span>
<span class="tooltip" data-tooltip="Processed securely by %s"><svg class="w-3 h-3"><use href="#ico-info"></use></svg></span>
</div>
`, HTMLEscapeString(n.RealmName), HTMLEscapeString(n.RealmName))

	// Track select elements that have been rendered
	renderedSelects := make(map[string]bool)
	lastDescID := ""

	// Render elements
	for i, elem := range n.Elements {
		if elem.GetError() != nil {
			fmt.Fprintf(w, "<!-- Error: %s -->\n", HTMLEscapeString(elem.GetError().Error()))
			continue
		}

		switch e := elem.(type) {
		case FormInput:
			r.renderInput(w, e, i, &lastDescID)
		case FormTextarea:
			r.renderTextarea(w, e, i, &lastDescID)
		case FormSelect:
			if !renderedSelects[e.Name] {
				r.renderSelect(w, n.Elements, e, i, &lastDescID)
				renderedSelects[e.Name] = true
			}
		}
	}

	// Submit button
	if len(n.Elements) > 0 {
		fmt.Fprintf(w, `<input type="submit" value="Submit to %s Realm" />`+"\n",
			HTMLEscapeString(n.RealmName))
	}

	fmt.Fprintln(w, "</form>")
	return ast.WalkContinue, nil
}

func (r *FormRenderer) renderInput(w util.BufWriter, e FormInput, idx int, lastDescID *string) {
	// Description
	if e.Description != "" {
		descID := fmt.Sprintf("desc_%s_%d", e.Name, idx)
		fmt.Fprintf(w, `<div id="%s" class="gno-form_description">%s</div>`+"\n",
			HTMLEscapeString(descID), HTMLEscapeString(e.Description))
		*lastDescID = descID
	}

	isSelectable := e.Type == "radio" || e.Type == "checkbox"

	if isSelectable {
		uniqueID := fmt.Sprintf("%s_%d", e.Name, idx)
		fmt.Fprintf(w, `<div class="gno-form_selectable">
<input type="%s" id="%s" name="%s" value="%s"`,
			HTMLEscapeString(e.Type),
			HTMLEscapeString(uniqueID),
			HTMLEscapeString(e.Name),
			HTMLEscapeString(e.Value))

		if *lastDescID != "" {
			fmt.Fprintf(w, ` aria-labelledby="%s"`, HTMLEscapeString(*lastDescID))
		}
		if e.Checked {
			fmt.Fprint(w, ` checked`)
		}
		fmt.Fprintln(w, ` />`)

		label := e.Value
		if e.Placeholder != "" {
			label += " - " + e.Placeholder
		}
		fmt.Fprintf(w, `<label for="%s"> %s </label>
</div>
`, HTMLEscapeString(uniqueID), HTMLEscapeString(label))
	} else {
		fmt.Fprintf(w, `<div class="gno-form_input"><label for="%s"> %s </label>
<input type="%s" id="%s" name="%s" placeholder="%s" />
</div>
`, HTMLEscapeString(e.Name), HTMLEscapeString(e.Placeholder),
			HTMLEscapeString(e.Type), HTMLEscapeString(e.Name),
			HTMLEscapeString(e.Name), HTMLEscapeString(e.Placeholder))
	}
}

func (r *FormRenderer) renderTextarea(w util.BufWriter, e FormTextarea, idx int, lastDescID *string) {
	if e.Description != "" {
		descID := fmt.Sprintf("desc_%s_%d", e.Name, idx)
		fmt.Fprintf(w, `<div id="%s" class="gno-form_description">%s</div>`+"\n",
			HTMLEscapeString(descID), HTMLEscapeString(e.Description))
		*lastDescID = descID
	}

	fmt.Fprintf(w, `<div class="gno-form_input"><label for="%s"> %s </label>
<textarea id="%s" name="%s" placeholder="%s" rows="%d"></textarea>
</div>
`, HTMLEscapeString(e.Name), HTMLEscapeString(e.Placeholder),
		HTMLEscapeString(e.Name), HTMLEscapeString(e.Name),
		HTMLEscapeString(e.Placeholder), e.Rows)
}

func (r *FormRenderer) renderSelect(w util.BufWriter, elements []FormElement, e FormSelect, idx int, lastDescID *string) {
	if e.Description != "" {
		descID := fmt.Sprintf("desc_%s_%d", e.Name, idx)
		fmt.Fprintf(w, `<div id="%s" class="gno-form_description">%s</div>`+"\n",
			HTMLEscapeString(descID), HTMLEscapeString(e.Description))
		*lastDescID = descID
	}

	label := titleCase(strings.ReplaceAll(e.Name, "_", " "))
	fmt.Fprintf(w, `<div class="gno-form_select"><label for="%s"> %s </label>
<select id="%s" name="%s"`,
		HTMLEscapeString(e.Name), HTMLEscapeString(label),
		HTMLEscapeString(e.Name), HTMLEscapeString(e.Name))

	if *lastDescID != "" {
		fmt.Fprintf(w, ` aria-labelledby="%s"`, HTMLEscapeString(*lastDescID))
	}
	fmt.Fprintln(w, `>`)

	article := GetWordArticle(label)
	fmt.Fprintf(w, `<option value="">Select %s %s</option>`+"\n",
		article, HTMLEscapeString(label))

	// Collect all options for this select
	for _, elem := range elements {
		if opt, ok := elem.(FormSelect); ok && opt.Name == e.Name {
			fmt.Fprintf(w, `<option value="%s"`, HTMLEscapeString(opt.Value))
			if opt.Selected {
				fmt.Fprint(w, ` selected`)
			}
			fmt.Fprintf(w, `>%s</option>`+"\n", HTMLEscapeString(opt.Value))
		}
	}

	fmt.Fprintln(w, `</select>
<svg class="w-4 h-4 absolute right-2 top-1/2 -translate-y-1/2 pointer-events-none"><use href="#ico-arrow"></use></svg>
</div>`)
}

// FormTransformer reorders form elements based on function signatures
type FormTransformer struct{}

func (t *FormTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			if formNode, ok := node.(*FormNode); ok && formNode.Exec != nil {
				t.reorderElements(formNode)
			}
		}
		return ast.WalkContinue, nil
	})
}

func (t *FormTransformer) reorderElements(node *FormNode) {
	var newElements []FormElement

	// If user provided any fields, error them all out
	if len(node.Elements) > 0 {
		for _, elem := range node.Elements {
			switch e := elem.(type) {
			case FormInput:
				e.Error = fmt.Errorf("manual fields not allowed when 'exec' is specified")
				newElements = append(newElements, e)
			case FormTextarea:
				e.Error = fmt.Errorf("manual fields not allowed when 'exec' is specified")
				newElements = append(newElements, e)
			case FormSelect:
				e.Error = fmt.Errorf("manual fields not allowed when 'exec' is specified")
				newElements = append(newElements, e)
			default:
				newElements = append(newElements, elem)
			}
		}
	}

	// Generate fields for all function parameters
	for _, param := range node.Exec.Params {
		newElements = append(newElements, t.createDefaultElement(param))
	}

	node.Elements = newElements
}

func (t *FormTransformer) createDefaultElement(param vm.NamedType) FormElement {
	placeholder := fmt.Sprintf("Enter %s", param.Name)

	switch param.Type {
	case "string":
		return FormInput{
			Name:        param.Name,
			Type:        "text",
			Placeholder: placeholder,
		}
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64":
		return FormInput{
			Name:        param.Name,
			Type:        "number",
			Placeholder: placeholder,
		}
	case "bool":
		return FormInput{
			Name:  param.Name,
			Type:  "checkbox",
			Value: "true",
		}
	default:
		return FormInput{
			Name:        param.Name,
			Type:        "text",
			Placeholder: fmt.Sprintf("Unsupported type: %s", param.Type),
			Error:       fmt.Errorf("%w '%s' for parameter '%s'", ErrFormUnsupportedType, param.Type, param.Name),
		}
	}
}

// FormExtension integrates forms into goldmark
type FormExtension struct{}

func (e *FormExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(util.Prioritized(NewFormParser(), 500)),
		parser.WithASTTransformers(util.Prioritized(&FormTransformer{}, 500)),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(util.Prioritized(NewFormRenderer(), 500)),
	)
}

// ExtForms is the public form extension instance
var ExtForms = &FormExtension{}

// Helper function for parsing form tags
func parseFormTag(line []byte) (html.Token, bool) {
	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return html.Token{}, false
	}
	toks, err := ParseHTMLTokens(bytes.NewReader(line))
	if err != nil || len(toks) != 1 {
		return html.Token{}, false
	}
	return toks[0], true
}
