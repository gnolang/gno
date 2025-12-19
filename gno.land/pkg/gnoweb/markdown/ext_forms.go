package markdown

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
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
	ErrFormInvalidTag       = errors.New("unexpected or invalid tag")
	ErrFormMissingName      = errors.New("missing 'name' attribute")
	ErrFormInvalidInputType = errors.New("invalid input type")
	ErrFormDuplicateName    = errors.New("name already used")
	ErrFormInvalidAttribute = errors.New("invalid attribute for input type")
	ErrFormMissingValue     = errors.New("missing 'value' attribute")
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
	Readonly    bool
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
	if e.Readonly {
		s += " (readonly=true)"
	}
	return s
}

// FormTextarea represents a textarea element
type FormTextarea struct {
	Name        string
	Placeholder string
	Rows        int
	Value       string
	Readonly    bool
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
	if e.Readonly {
		s += " (readonly=true)"
	}
	return s
}

// FormSelect represents a select option
type FormSelect struct {
	Name        string
	Value       string
	Selected    bool
	Readonly    bool
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
	if e.Readonly {
		s += " (readonly=true)"
	}
	return s
}

// FormNode represents a form in the AST
type FormNode struct {
	ast.BaseBlock
	Elements   []FormElement
	ExecFunc   string // Function name for exec attribute
	RenderPath string
	RealmName  string
	Domain     string
	ChainId    string
	Remote     string
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

	// Get ChainId, Remote, and Domain from context
	if chainId, ok := getChainIdFromContext(pc); ok {
		node.ChainId = chainId
	}
	if remote, ok := getRemoteFromContext(pc); ok {
		node.Remote = remote
	}
	if domain, ok := getDomainFromContext(pc); ok {
		node.Domain = domain
	}

	// Handle exec attribute
	if exec, ok := ExtractAttr(tok.Attr, "exec"); ok {
		node.ExecFunc = exec
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
	input.Readonly = attrs["readonly"] == "true"

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
		case "value":
			textarea.Value = strings.NewReplacer("\\n", "\n", "\\t", "\t").Replace(strings.TrimSpace(attr.Val))
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
		case "readonly":
			textarea.Readonly = strings.TrimSpace(attr.Val) == "true"
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
		case "readonly":
			sel.Readonly = strings.TrimSpace(attr.Val) == "true"
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
	fmt.Fprintf(w, `<form class="gno-form" method="post" action="%s" autocomplete="off" spellcheck="false"`, HTMLEscapeString(action))
	if n.ExecFunc != "" {
		fmt.Fprintf(w, ` data-controller="form-exec"`)
	}
	fmt.Fprintf(w, `>`+"\n")
	headerLabel := "Form"
	if n.ExecFunc != "" {
		headerLabel = fmt.Sprintf("Exec: %s", HTMLEscapeString(titleCase(n.ExecFunc)))
	}
	fmt.Fprintf(w, `<div class="gno-form_header">
<span><span class="font-bold">%s</span> %s</span>
<span class="tooltip" data-tooltip="Processed securely by %s"><svg class="w-3 h-3"><use href="#ico-info"></use></svg></span>
</div>
`, HTMLEscapeString(n.RealmName), headerLabel, HTMLEscapeString(n.RealmName))

	if n.ExecFunc != "" {
		fmt.Fprintf(w, `<div data-controller="action-function" data-action-function-name-value="%s">`+"\n", HTMLEscapeString(n.ExecFunc))
	}

	// Track select elements that have been rendered
	renderedSelects := make(map[string]bool)
	lastDescID := ""

	// Render elements
	isExec := n.ExecFunc != ""
	for i, elem := range n.Elements {
		if elem.GetError() != nil {
			fmt.Fprintf(w, "<!-- Error: %s -->\n", HTMLEscapeString(elem.GetError().Error()))
			continue
		}

		switch e := elem.(type) {
		case FormInput:
			r.renderInput(w, e, i, &lastDescID, isExec)
		case FormTextarea:
			r.renderTextarea(w, e, i, &lastDescID, isExec)
		case FormSelect:
			if !renderedSelects[e.Name] {
				r.renderSelect(w, n.Elements, e, i, &lastDescID, isExec)
				renderedSelects[e.Name] = true
			}
		}
	}

	// Submit button
	if len(n.Elements) > 0 {
		if n.ExecFunc != "" {
			fmt.Fprintf(w, `<div class="gno-form_input"><input type="submit" value="Submit (%s Function)" /></div>`+"\n",
				HTMLEscapeString(n.ExecFunc))
		} else {
			fmt.Fprintf(w, `<div class="gno-form_input"><input type="submit" value="Submit to %s Realm" /></div>`+"\n",
				HTMLEscapeString(n.RealmName))
		}
	}

	// Add command block if we have an exec function
	if n.ExecFunc != "" {
		fmt.Fprintf(w, `<div class="command u-hidden" data-form-exec-target="command">`)
		// Add mode and address controls if we have an exec function
		fmt.Fprintf(w, `<div data-controller="action-header" class="c-between">
  <span class="title">Command</span>
  <div class="c-inline">
    <div class="b-input">
      <select data-action-header-target="mode" data-action="change->action-header#updateMode">
        <option value="secure" selected="selected">Mode: Full Security</option>
        <option value="fast">Mode: Fast</option>
      </select>
      <svg><use href="#ico-arrow-down"></use></svg>
    </div>
    <div class="b-input">
      <label for="form-address-%s">Address</label>
      <input type="text" data-action-header-target="address" data-action="input->action-header#updateAddress" id="form-address-%s" class="u-font-mono" placeholder="ADDRESS" />
    </div>
  </div>
</div>`, HTMLEscapeString(n.ExecFunc), HTMLEscapeString(n.ExecFunc))
		r.renderCommandBlock(w, n)
		fmt.Fprintln(w, `</div>`)
		fmt.Fprintln(w, `</div>`)
	}

	fmt.Fprintln(w, "</form>")

	return ast.WalkContinue, nil
}

func (r *FormRenderer) renderCommandBlock(w util.BufWriter, n *FormNode) {
	// Use default values if not set
	chainId := n.ChainId
	if chainId == "" {
		chainId = "dev"
	}
	remote := n.Remote
	if remote == "" {
		remote = "127.0.0.1:26657"
	}

	// Extract unique parameter names (preserving order and avoiding duplicates)
	seen := make(map[string]bool)
	var paramNames []string
	for _, elem := range n.Elements {
		if name := elem.GetName(); name != "" && !seen[name] && elem.GetError() == nil {
			paramNames = append(paramNames, name)
			seen[name] = true
		}
	}

	// Prepare data for the command template
	// Build PkgPath with domain (like the action page does)
	pkgPath := n.Domain + n.RealmName
	data := components.CommandData{
		FuncName:   n.ExecFunc,
		PkgPath:    pkgPath,
		ParamNames: paramNames,
		ChainId:    chainId,
		Remote:     remote,
	}

	// Create and render the template component
	comp := components.NewTemplateComponent("ui/command", data)
	if err := comp.Render(w); err != nil {
		fmt.Fprintf(w, "<!-- Error rendering command block: %s -->\n", HTMLEscapeString(err.Error()))
	}
}

func (r *FormRenderer) renderInput(w util.BufWriter, e FormInput, idx int, lastDescID *string, isExec bool) {
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
		if e.Readonly {
			fmt.Fprint(w, ` disabled`)
		}
		if isExec {
			fmt.Fprintf(w, ` data-action-function-target="param-input" data-action="change->action-function#updateAllArgs" data-action-function-param-value="%s"`, HTMLEscapeString(e.Name))
		}

		fmt.Fprintln(w, ` />`)

		label := e.Value
		if e.Placeholder != "" {
			label += " - " + e.Placeholder
		}
		readonlyBadge := ""
		if e.Readonly {
			readonlyBadge = `<span class="gno-form_readonly-badge">(readonly)</span>`
		}
		fmt.Fprintf(w, `<label for="%s"> %s %s</label>
</div>
`, HTMLEscapeString(uniqueID), HTMLEscapeString(label), readonlyBadge)
	} else {
		readonlyBadge := ""
		if e.Readonly {
			readonlyBadge = `<span class="gno-form_readonly-badge">(readonly)</span>`
		}
		fmt.Fprintf(w, `<div class="gno-form_input"><label for="%s"> %s %s</label>
<input type="%s" id="%s" name="%s" placeholder="%s"`,
			HTMLEscapeString(e.Name), HTMLEscapeString(e.Placeholder), readonlyBadge,
			HTMLEscapeString(e.Type), HTMLEscapeString(e.Name),
			HTMLEscapeString(e.Name), HTMLEscapeString(e.Placeholder))
		if e.Value != "" {
			fmt.Fprintf(w, ` value="%s"`, HTMLEscapeString(e.Value))
		}
		if e.Readonly {
			fmt.Fprint(w, ` readonly`)
		}
		if isExec {
			fmt.Fprintf(w, ` data-action-function-target="param-input" data-action="input->action-function#updateAllArgs" data-action-function-param-value="%s"`, HTMLEscapeString(e.Name))
		}
		fmt.Fprintln(w, ` />
</div>`)
	}
}

func (r *FormRenderer) renderTextarea(w util.BufWriter, e FormTextarea, idx int, lastDescID *string, isExec bool) {
	if e.Description != "" {
		descID := fmt.Sprintf("desc_%s_%d", e.Name, idx)
		fmt.Fprintf(w, `<div id="%s" class="gno-form_description">%s</div>`+"\n",
			HTMLEscapeString(descID), HTMLEscapeString(e.Description))
		*lastDescID = descID
	}

	readonlyBadge := ""
	if e.Readonly {
		readonlyBadge = `<span class="gno-form_readonly-badge">(readonly)</span>`
	}

	fmt.Fprintf(w, `<div class="gno-form_input"><label for="%s"> %s %s</label>
<textarea id="%s" name="%s" placeholder="%s" rows="%d"`,
		HTMLEscapeString(e.Name), HTMLEscapeString(e.Placeholder), readonlyBadge,
		HTMLEscapeString(e.Name), HTMLEscapeString(e.Name),
		HTMLEscapeString(e.Placeholder), e.Rows)
	if e.Readonly {
		fmt.Fprint(w, ` readonly`)
	}
	if isExec {
		fmt.Fprintf(w, ` data-action-function-target="param-input" data-action="input->action-function#updateAllArgs" data-action-function-param-value="%s"`, HTMLEscapeString(e.Name))
	}
	fmt.Fprintf(w, `>%s</textarea>
</div>
`, HTMLEscapeString(e.Value))
}

func (r *FormRenderer) renderSelect(w util.BufWriter, elements []FormElement, e FormSelect, idx int, lastDescID *string, isExec bool) {
	if e.Description != "" {
		descID := fmt.Sprintf("desc_%s_%d", e.Name, idx)
		fmt.Fprintf(w, `<div id="%s" class="gno-form_description">%s</div>`+"\n",
			HTMLEscapeString(descID), HTMLEscapeString(e.Description))
		*lastDescID = descID
	}

	label := titleCase(strings.ReplaceAll(e.Name, "_", " "))
	readonlyBadge := ""
	if e.Readonly {
		readonlyBadge = `<span class="gno-form_readonly-badge">(readonly)</span>`
	}
	fmt.Fprintf(w, `<div class="gno-form_select"><label for="%s"> %s %s</label>
<select id="%s" name="%s"`,
		HTMLEscapeString(e.Name), HTMLEscapeString(label), readonlyBadge,
		HTMLEscapeString(e.Name), HTMLEscapeString(e.Name))

	if *lastDescID != "" {
		fmt.Fprintf(w, ` aria-labelledby="%s"`, HTMLEscapeString(*lastDescID))
	}
	if e.Readonly {
		fmt.Fprint(w, ` disabled`)
	}
	if isExec {
		fmt.Fprintf(w, ` data-action-function-target="param-input" data-action="change->action-function#updateAllArgs" data-action-function-param-value="%s"`, HTMLEscapeString(e.Name))
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

// FormExtension integrates forms into goldmark
type FormExtension struct{}

func (e *FormExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(util.Prioritized(NewFormParser(), 500)),
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
