package template

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"gno.land/p/moul/md"
	"gno.land/p/moul/typeutil"
)

// Enable/disable debug logs
var debug bool = false

func log(args ...interface{}) {
	if !debug {
		return
	}
	var s string
	for i, arg := range args {
		if i > 0 {
			s += " "
		}
		s += typeutil.ToString(arg)
	}
	println(s)
}

// Context kinds
const (
	kindNone = iota
	kindRange
	kindIf
	kindElse
)

// context represents the current "range" or "if" scope
type context struct {
	parent       *context
	currentRange []interface{}
	rangeMap     map[string]interface{}
	rangeIndex   int
	kind         int
}

// Renderer is our main template struct
type Renderer struct {
	helpers     map[string]Helper
	data        map[string]interface{}
	context     *context
	lastContent string
}

// NewRenderer creates a new template renderer with default helpers
func NewRenderer() *Renderer {
	r := &Renderer{
		helpers: make(map[string]Helper),
		data:    make(map[string]interface{}),
	}
	// Register markdown helpers
	r.helpers["H1"] = SingleArgHelper{fn: md.H1}
	r.helpers["H2"] = SingleArgHelper{fn: md.H2}
	r.helpers["H3"] = SingleArgHelper{fn: md.H3}
	r.helpers["H4"] = SingleArgHelper{fn: md.H4}
	r.helpers["H5"] = SingleArgHelper{fn: md.H5}
	r.helpers["H6"] = SingleArgHelper{fn: md.H6}
	r.helpers["Bold"] = SingleArgHelper{fn: md.Bold}
	r.helpers["Italic"] = SingleArgHelper{fn: md.Italic}
	r.helpers["Strikethrough"] = SingleArgHelper{fn: md.Strikethrough}
	r.helpers["InlineCode"] = SingleArgHelper{fn: md.InlineCode}
	r.helpers["BulletItem"] = SingleArgHelper{fn: md.BulletItem}
	r.helpers["TodoItem"] = TwoArgStringBoolHelper{fn: md.TodoItem}
	r.helpers["CodeBlock"] = SingleArgHelper{fn: md.CodeBlock}
	r.helpers["LanguageCodeBlock"] = TwoArgHelper{fn: md.LanguageCodeBlock}
	r.helpers["Link"] = TwoArgHelper{fn: md.Link}
	r.helpers["Image"] = TwoArgHelper{fn: md.Image}
	r.helpers["InlineImageWithLink"] = ThreeArgHelper{fn: md.InlineImageWithLink}
	r.helpers["Blockquote"] = SingleArgHelper{fn: md.Blockquote}
	r.helpers["HorizontalRule"] = SingleArgHelper{fn: func(string) string { return md.HorizontalRule() }}
	r.helpers["Footnote"] = TwoArgHelper{fn: md.Footnote}
	r.helpers["EscapeText"] = SingleArgHelper{fn: md.EscapeText}
	r.helpers["Paragraph"] = SingleArgHelper{fn: md.Paragraph}
	r.helpers["CollapsibleSection"] = TwoArgHelper{fn: md.CollapsibleSection}

	// Register other helpers
	r.helpers["concat"] = VarArgHelper{fn: concatHelper}
	return r
}

// Render processes the passed-in template string and returns the rendered result
func (r *Renderer) Render(template string, data map[string]interface{}) string {
	if data != nil {
		r.data = data
	}
	r.lastContent = template

	var result strings.Builder
	i := 0
	for i < len(template) {
		// Look for a "{{"
		if i+1 < len(template) && template[i] == '{' && template[i+1] == '{' {
			// Write out anything before the "{{"
			placeholderStart := i
			result.WriteString(template[:placeholderStart])

			// Skip "{{"
			i += 2

			// Parse out the placeholder content until matching "}}"
			inner, trimLeft, trimRight, endPos := r.parsePlaceholder(template, i)
			if endPos < 0 {
				// No matching }} found; write the rest as-is and break
				result.WriteString(template[placeholderStart:])
				break
			}
			// Move i to after "}}"
			i = endPos + 2

			// Trim left (remove trailing spaces in result)
			if trimLeft {
				resultString := result.String()
				result.Reset()
				result.WriteString(strings.TrimRightFunc(resultString, unicode.IsSpace))
			}

			// Process that placeholder
			output, err := r.processPlaceholder(inner)
			if err != nil {
				// In debug mode we might display these errors, or just inline them
				output = fmt.Sprintf("[error: %v]", err)
			}
			result.WriteString(output)

			// Trim right from the source template (i.e., skip spaces in template)
			if trimRight {
				for i < len(template) && unicode.IsSpace(rune(template[i])) {
					i++
				}
			}

			// The chunk is fully handled, so shift template to skip placeholder
			template = template[i:]
			i = 0 // reset to parse from beginning of remainder

		} else {
			// Normal character, just move forward
			i++
		}
	}

	// If there's any remainder after last placeholder, write it out
	result.WriteString(template)
	return result.String()
}

// parsePlaceholder extracts the placeholder content between "{{" and "}}",
// also checks for the optional trim markers "-}}", "{{-".
func (r *Renderer) parsePlaceholder(template string, start int) (inner string, trimLeft bool, trimRight bool, endPos int) {
	// We'll parse until we match the next "}}" (noting we might skip nested "{{...}}")
	depth := 1
	j := start
	for j < len(template)-1 {
		if template[j] == '{' && template[j+1] == '{' {
			depth++
			j += 2
		} else if template[j] == '}' && template[j+1] == '}' {
			depth--
			if depth == 0 {
				break
			}
			j += 2
		} else {
			j++
		}
	}
	if depth != 0 {
		return "", false, false, -1 // no matching "}}"
	}

	raw := template[start:j]

	// Check for trim markers
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "-") {
		trimLeft = true
		raw = strings.TrimPrefix(raw, "-")
		raw = strings.TrimLeftFunc(raw, unicode.IsSpace)
	}
	if strings.HasSuffix(raw, "-") {
		trimRight = true
		raw = strings.TrimSuffix(raw, "-")
		raw = strings.TrimRightFunc(raw, unicode.IsSpace)
	}

	return strings.TrimSpace(raw), trimLeft, trimRight, j
}

// processPlaceholder decides how to handle the placeholder logic: variable, range, if, helper, etc.
func (r *Renderer) processPlaceholder(content string) (string, error) {
	parts := splitParts(content)
	if len(parts) == 0 {
		return "", nil
	}
	// If the content starts with "." it might be a variable
	if strings.HasPrefix(parts[0], ".") {
		val, ok := r.getVar(parts[0])
		if ok {
			return typeutil.ToString(val), nil
		}
		return "", nil
	}

	switch parts[0] {
	case "range":
		if len(parts) < 2 {
			return "", errors.New("range requires a variable")
		}
		return r.handleRange(parts[1:])
	case "end":
		// "end" should be handled inside handleRange or handleIf, so normally we wouldn't get here.
		return "", nil
	case "if":
		if len(parts) < 2 {
			return "", fmt.Errorf("if requires a condition")
		}
		return r.handleIf(parts[1:])
	default:
		// Possibly a helper
		if helper, ok := r.helpers[parts[0]]; ok {
			return helper.Execute(parts[1:], r.data), nil
		}
		// Otherwise just treat as plain text
		return strings.Join(parts, " "), nil
	}
}

// handleRange scans forward in r.lastContent to find the matching {{ end }} block content,
// then iterates over each item in the slice, calling subrender.
func (r *Renderer) handleRange(exprParts []string) (string, error) {
	val, ok := r.getVar("." + strings.TrimPrefix(exprParts[0], ".")) // safe fix in case user used .Var or Var
	if !ok {
		return "", nil
	}

	// We want to collect the content between "{{ range ... }}" and the matching "{{ end }}"
	rangeBlock, err := r.extractBlock("range "+strings.Join(exprParts, " "), "end")
	if err != nil {
		return "", err
	}

	var rangeData []interface{}
	switch v := val.(type) {
	case []interface{}:
		rangeData = v
	case []map[string]interface{}:
		rangeData = make([]interface{}, len(v))
		for i, m := range v {
			rangeData[i] = m
		}
	default:
		if slice := typeutil.ToInterfaceSlice(v); slice != nil {
			rangeData = slice
		} else {
			return "", fmt.Errorf("cannot range over %T", val)
		}
	}

	var result strings.Builder
	for i, item := range rangeData {
		newCtx := &context{
			parent:       r.context,
			kind:         kindRange,
			currentRange: rangeData,
			rangeIndex:   i,
		}
		if m, isMap := item.(map[string]interface{}); isMap {
			newCtx.rangeMap = m
		}

		subRenderer := &Renderer{
			helpers: r.helpers,
			data:    r.data,
			context: newCtx,
		}
		subResult := subRenderer.Render(rangeBlock, nil)
		result.WriteString(subResult)
	}

	return result.String(), nil
}

// handleIf scans forward in r.lastContent to find the matching {{ end }} (and possible {{ else }}),
// renders the relevant block if condition is true or false.
func (r *Renderer) handleIf(exprParts []string) (string, error) {
	condVal, ok := r.getVar("." + strings.TrimPrefix(exprParts[0], "."))
	if !ok {
		condVal = false
	}
	condition := typeutil.ToBool(condVal)
	fullBlock, err := r.extractBlock("if "+strings.Join(exprParts, " "), "end")
	if err != nil {
		return "", err
	}

	// Try to find an {{ else }} within that block to split it
	elseIdx := -1
	depth := 0
	for i := 0; i+2 < len(fullBlock); i++ {
		if strings.HasPrefix(fullBlock[i:], "{{") {
			if strings.HasPrefix(fullBlock[i+2:], " if ") {
				depth++
			} else if strings.HasPrefix(fullBlock[i+2:], " end") {
				// match: reduce depth
				depth--
			} else if depth == 0 && strings.HasPrefix(fullBlock[i+2:], " else") {
				elseIdx = i
				break
			}
		}
	}

	var ifBlock, elseBlock string
	if elseIdx >= 0 {
		ifBlock = fullBlock[:elseIdx]
		// skip the "{{ else }}" placeholder
		endOfElse := elseIdx
		for endOfElse < len(fullBlock) && fullBlock[endOfElse:endOfElse+2] != "}}" {
			endOfElse++
		}
		endOfElse += 2
		elseBlock = fullBlock[endOfElse:]
	} else {
		ifBlock = fullBlock
	}

	newCtx := &context{
		parent: r.context,
		kind:   kindIf,
	}
	subRenderer := &Renderer{
		helpers: r.helpers,
		data:    r.data,
		context: newCtx,
	}

	if condition {
		return subRenderer.Render(ifBlock, nil), nil
	} else {
		return subRenderer.Render(elseBlock, nil), nil
	}
}

// extractBlock finds the text between "{{ range/if expr }}" and the matching "{{ end }}".
// We rely on r.lastContent to find the entire block so that nested placeholders can be found.
func (r *Renderer) extractBlock(startExpr, endExpr string) (string, error) {
	// Look for the exact "{{ startExpr }}" and keep track of nested blocks until we find the matching "{{ endExpr }}"
	startMarker1 := "{{ " + startExpr + " }}"
	startMarker2 := "{{" + startExpr + "}}" // fallback no spaces
	startIdx := strings.Index(r.lastContent, startMarker1)
	if startIdx < 0 {
		startIdx = strings.Index(r.lastContent, startMarker2)
		if startIdx < 0 {
			return "", fmt.Errorf("could not find start marker for %q", startExpr)
		}
	}
	blockStart := startIdx + len(startMarker1)
	if strings.Contains(r.lastContent[startIdx:startIdx+len(startMarker1)], startExpr) == false {
		// If we matched the fallback startMarker2, adjust blockStart
		blockStart = startIdx + len(startMarker2)
	}

	depth := 1
	i := blockStart
	var block strings.Builder

	for i < len(r.lastContent)-1 {
		// if we see "{{"
		if r.lastContent[i] == '{' && r.lastContent[i+1] == '{' {
			// check what follows
			innerStart := i + 2
			// find matching "}}"
			_, _, _, matchPos := r.parsePlaceholder(r.lastContent, innerStart)
			if matchPos < 0 {
				// no matching "}}", break
				break
			}
			innerContent := strings.TrimSpace(r.lastContent[innerStart:matchPos])

			// check if it's a nested start
			if strings.HasPrefix(innerContent, "range ") || strings.HasPrefix(innerContent, "if ") {
				depth++
			} else if strings.HasPrefix(innerContent, endExpr) {
				depth--
				if depth == 0 {
					// done, skip past that end
					i = matchPos + 2
					break
				}
			}
			block.WriteString(r.lastContent[i : matchPos+2])
			i = matchPos + 2
		} else {
			block.WriteByte(r.lastContent[i])
			i++
		}
	}

	// Now we have the block content (minus the outside markers)
	return block.String(), nil
}

// splitParts splits a template expression into parts, respecting quoted strings
func splitParts(s string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"':
			inQuote = !inQuote
			current.WriteByte(s[i])
		case ' ':
			if !inQuote {
				if current.Len() > 0 {
					parts = append(parts, current.String())
					current.Reset()
				}
			} else {
				current.WriteByte(s[i])
			}
		default:
			current.WriteByte(s[i])
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

// getVar retrieves a variable from context or global data
func (r *Renderer) getVar(name string) (interface{}, bool) {
	if r.context != nil && r.context.kind == kindRange {
		// Inside a range block
		if name == "." {
			if r.context.currentRange != nil && r.context.rangeIndex < len(r.context.currentRange) {
				return r.context.currentRange[r.context.rangeIndex], true
			}
		} else if strings.HasPrefix(name, ".") {
			name = strings.TrimPrefix(name, ".")
			// if we have a rangeMap
			if r.context.rangeMap != nil {
				if val, ok := r.context.rangeMap[name]; ok {
					return val, true
				}
			}
			// or if the item is a map
			if r.context.currentRange != nil && r.context.rangeIndex < len(r.context.currentRange) {
				if item, ok := r.context.currentRange[r.context.rangeIndex].(map[string]interface{}); ok {
					if val, exists := item[name]; exists {
						return val, true
					}
				}
			}
		}
	}
	// fallback: look in global data
	return r.lookupVar(strings.TrimPrefix(name, "."))
}

// lookupVar supports dot-notation "nested.key"
func (r *Renderer) lookupVar(name string) (interface{}, bool) {
	parts := strings.Split(name, ".")
	cur := r.data
	for i, part := range parts {
		if i == len(parts)-1 {
			val, ok := cur[part]
			return val, ok
		}
		next, ok := cur[part]
		if !ok {
			return nil, false
		}
		nextMap, isMap := next.(map[string]interface{})
		if !isMap {
			return nil, false
		}
		cur = nextMap
	}
	return nil, false
}

// Helpers

// Helper is an interface for template helpers
type Helper interface {
	Execute(args []string, data map[string]interface{}) string
}

type SingleArgHelper struct{ fn func(string) string }

func (h SingleArgHelper) Execute(args []string, data map[string]interface{}) string {
	if len(args) < 1 {
		return "[error: missing argument]"
	}
	arg := strings.Join(args, " ")
	// If it has outer quotes, strip them
	if len(arg) >= 2 && arg[0] == '"' && arg[len(arg)-1] == '"' {
		arg = arg[1 : len(arg)-1]
	}
	return h.fn(strings.TrimSpace(arg))
}

type VarArgHelper struct{ fn func([]string) string }

func (h VarArgHelper) Execute(args []string, data map[string]interface{}) string {
	processed := make([]string, len(args))
	for i, arg := range args {
		arg = strings.TrimSpace(arg)
		if strings.HasPrefix(arg, ".") {
			if val, ok := data[strings.TrimPrefix(arg, ".")]; ok {
				processed[i] = typeutil.ToString(val)
				continue
			}
		}
		// quoted?
		if len(arg) >= 2 && arg[0] == '"' && arg[len(arg)-1] == '"' {
			arg = arg[1 : len(arg)-1]
		}
		processed[i] = arg
	}
	return h.fn(processed)
}

type TwoArgHelper struct{ fn func(string, string) string }

func (h TwoArgHelper) Execute(args []string, data map[string]interface{}) string {
	if len(args) < 2 {
		return "[error: requires two arguments]"
	}
	arg1 := trimQuotes(args[0])
	arg2 := trimQuotes(args[1])
	return h.fn(arg1, arg2)
}

type TwoArgStringBoolHelper struct{ fn func(string, bool) string }

func (h TwoArgStringBoolHelper) Execute(args []string, data map[string]interface{}) string {
	if len(args) < 2 {
		return "[error: requires two arguments]"
	}
	arg1 := trimQuotes(args[0])
	arg2 := strings.ToLower(args[1]) == "true"
	return h.fn(arg1, arg2)
}

type ThreeArgHelper struct {
	fn func(string, string, string) string
}

func (h ThreeArgHelper) Execute(args []string, data map[string]interface{}) string {
	if len(args) < 3 {
		return "[error: requires three arguments]"
	}
	arg1 := trimQuotes(args[0])
	arg2 := trimQuotes(args[1])
	arg3 := trimQuotes(args[2])
	return h.fn(arg1, arg2, arg3)
}

// Simple helper for string concatenation
func concatHelper(args []string) string {
	return strings.Join(args, "")
}

func trimQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	return s
}
