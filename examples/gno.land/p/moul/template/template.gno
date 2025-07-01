package template

import (
	"errors"
	"fmt"
	"strconv"
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
	r.helpers["HorizontalRule"] = ZeroArgHelper{fn: md.HorizontalRule}
	r.helpers["Footnote"] = TwoArgHelper{fn: md.Footnote}
	r.helpers["EscapeText"] = SingleArgHelper{fn: md.EscapeText}
	r.helpers["Paragraph"] = SingleArgHelper{fn: md.Paragraph}
	r.helpers["CollapsibleSection"] = TwoArgHelper{fn: md.CollapsibleSection}

	// Register other helpers
	r.helpers["concat"] = VarArgHelper{fn: concatHelper}
	r.helpers["printf"] = VarArgHelper{fn: printfHelper}
	r.helpers["index"] = TwoArgHelper{fn: indexHelper}
	r.helpers["string"] = SingleArgHelper{fn: stringHelper}
	r.helpers["trim"] = SingleArgHelper{fn: trimHelper}
	return r
}

// Render processes the passed-in template string and returns the rendered result
func (r *Renderer) Render(template string, data map[string]interface{}) string {
	if data != nil {
		r.data = data
	}
	r.lastContent = template

	var result strings.Builder
	originalTemplate := template
	templateOffset := 0 // Track how much of the original template we've consumed
	
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

			// Update lastContent to reflect current position for block extraction
			r.lastContent = originalTemplate[templateOffset:]

			// Process that placeholder
			log("Processing placeholder:", inner)
			output, err, consumed := r.processPlaceholder(inner)
			if err != nil {
				// In debug mode we might display these errors, or just inline them
				output = fmt.Sprintf("[error: %v]", err)
			}
			log("Placeholder output:", output, "consumed:", consumed)
			result.WriteString(output)
			
			// If this placeholder consumed additional content (like range/if blocks),
			// skip that content in our template
			if consumed > 0 {
				// consumed is relative to the current lastContent position
				remainingAfterConsumption := originalTemplate[templateOffset+consumed:]
				template = remainingAfterConsumption
				templateOffset += consumed
				i = 0 // Reset position since template changed
				continue
			}

			// Trim right from the source template (i.e., skip spaces in template)
			if trimRight {
				for i < len(template) && unicode.IsSpace(rune(template[i])) {
					i++
				}
			}

			// The chunk is fully handled, so shift template to skip placeholder
			templateOffset += i
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
// Returns (output, error, consumedBytes) where consumedBytes indicates how much additional content was consumed
func (r *Renderer) processPlaceholder(content string) (string, error, int) {
	parts := splitParts(content)
	if len(parts) == 0 {
		return "", nil, 0
	}
	// If the content starts with "." it might be a variable
	if strings.HasPrefix(parts[0], ".") {
		val, ok := r.getVar(parts[0])
		if ok {
			return typeutil.ToString(val), nil, 0
		}
		return "", nil, 0
	}

	switch parts[0] {
	case "range":
		if len(parts) < 2 {
			return "", errors.New("range requires a variable"), 0
		}
		output, consumed := r.handleRangeWithConsumption(parts[1:])
		return output, nil, consumed
	case "end":
		// "end" should be handled inside handleRange or handleIf, so normally we wouldn't get here.
		return "", nil, 0
	case "if":
		if len(parts) < 2 {
			return "", fmt.Errorf("if requires a condition"), 0
		}
		output, consumed := r.handleIfWithConsumption(parts[1:])
		return output, nil, consumed
	case "else":
		// "else" should be handled inside handleIf, so normally we wouldn't get here.
		return "", nil, 0
	default:
		// Possibly a helper
		if helper, ok := r.helpers[parts[0]]; ok {
			// Special handling for index helper - try advanced indexing first
			if parts[0] == "index" && len(parts[1:]) >= 2 {
				firstArg := strings.TrimSpace(parts[1])
				if strings.HasPrefix(firstArg, ".") {
					// This looks like a variable access - use advanced indexing
					result, err, consumed := r.handleIndexHelper(parts[1:])
					if err == nil {
						return result, nil, consumed
					}
					// Fall back to simple index helper if advanced fails
				}
			}
			
			// Resolve variables and nested function calls in helper arguments
			resolvedArgs := make([]string, len(parts[1:]))
			for i, arg := range parts[1:] {
				arg = strings.TrimSpace(arg)
				
				// Handle nested function calls like (printf ...)
				if strings.HasPrefix(arg, "(") && strings.HasSuffix(arg, ")") {
					innerExpr := arg[1 : len(arg)-1] // Remove outer parentheses
					innerOutput, _, _ := r.processPlaceholder(innerExpr)
					resolvedArgs[i] = innerOutput
				} else if strings.HasPrefix(arg, ".") {
					if val, ok := r.getVar(arg); ok {
						resolvedArgs[i] = typeutil.ToString(val)
					} else {
						resolvedArgs[i] = arg
					}
				} else {
					// Check if the argument contains nested template expressions
					if strings.Contains(arg, "{{") && strings.Contains(arg, "}}") {
						// Process nested templates in the argument
						nestedOutput := r.Render(arg, nil)
						resolvedArgs[i] = nestedOutput
					} else {
						resolvedArgs[i] = arg
					}
				}
			}
			return helper.Execute(resolvedArgs, r.data), nil, 0
		}
		// Otherwise just treat as plain text
		return strings.Join(parts, " "), nil, 0
	}
}

// handleRangeWithConsumption processes a range block and returns output + consumed bytes
func (r *Renderer) handleRangeWithConsumption(exprParts []string) (string, int) {
	val, ok := r.getVar("." + strings.TrimPrefix(exprParts[0], "."))
	if !ok {
		return "", 0
	}

	// Find the range block and calculate how many bytes to consume
	rangeBlock, consumed, err := r.extractBlockWithConsumption("range "+strings.Join(exprParts, " "), "end")
	if err != nil {
		return fmt.Sprintf("[error: %v]", err), 0
	}
	log("Range block extracted:", rangeBlock, "consumed:", consumed)

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
			return fmt.Sprintf("[error: cannot range over %T]", val), 0
		}
	}

	var result strings.Builder
	log("Range data:", rangeData, "length:", len(rangeData))
	for i, item := range rangeData {
		log("Range iteration", i, "item:", item)
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
		log("Sub result:", subResult)
		result.WriteString(subResult)
	}

	return result.String(), consumed
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
	log("Range block extracted:", rangeBlock)

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
	log("Range data:", rangeData, "length:", len(rangeData))
	for i, item := range rangeData {
		log("Range iteration", i, "item:", item)
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
		log("Sub result:", subResult)
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

// extractBlock finds the text between the current position and the matching "{{ end }}".
func (r *Renderer) extractBlock(startExpr, endExpr string) (string, error) {
	// Find the opening placeholder in the template
	var startIdx int = -1
	var blockStart int
	
	// Try to find any {{ range ... }} or {{ if ... }} that matches our start pattern
	i := 0
	for i < len(r.lastContent)-1 {
		if r.lastContent[i] == '{' && r.lastContent[i+1] == '{' {
			innerStart := i + 2
			innerContent, _, _, matchPos := r.parsePlaceholder(r.lastContent, innerStart)
			if matchPos < 0 {
				i++
				continue
			}
			
			if strings.HasPrefix(innerContent, strings.Split(startExpr, " ")[0]) {
				startIdx = i
				blockStart = matchPos + 2
				break
			}
			i = matchPos + 2
		} else {
			i++
		}
	}
	
	if startIdx < 0 {
		return "", fmt.Errorf("could not find start marker for %q", startExpr)
	}

	depth := 1
	i = blockStart
	var block strings.Builder

	for i < len(r.lastContent)-1 {
		if r.lastContent[i] == '{' && r.lastContent[i+1] == '{' {
			innerStart := i + 2
			innerContent, _, _, matchPos := r.parsePlaceholder(r.lastContent, innerStart)
			if matchPos < 0 {
				break
			}

			// check if it's a nested start
			if strings.HasPrefix(innerContent, "range ") || strings.HasPrefix(innerContent, "if ") {
				depth++
			} else if strings.HasPrefix(innerContent, endExpr) {
				depth--
				if depth == 0 {
					// done, don't include the end marker
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

	return block.String(), nil
}

// extractBlockWithConsumption finds the block content and returns how many bytes to consume
func (r *Renderer) extractBlockWithConsumption(startExpr, endExpr string) (string, int, error) {
	template := r.lastContent
	
	// Find the opening placeholder in the template
	var startIdx int = -1
	var blockStart int
	
	// Try to find any {{ range ... }} or {{ if ... }} that matches our start pattern
	i := 0
	for i < len(template)-1 {
		if template[i] == '{' && template[i+1] == '{' {
			innerStart := i + 2
			innerContent, _, _, matchPos := r.parsePlaceholder(template, innerStart)
			if matchPos < 0 {
				i++
				continue
			}
			
			if strings.HasPrefix(innerContent, strings.Split(startExpr, " ")[0]) {
				startIdx = i
				blockStart = matchPos + 2
				break
			}
			i = matchPos + 2
		} else {
			i++
		}
	}
	
	if startIdx < 0 {
		return "", 0, fmt.Errorf("could not find start marker for %q", startExpr)
	}

	depth := 1
	i = blockStart
	var block strings.Builder
	var endPos int

	for i < len(template)-1 {
		if template[i] == '{' && template[i+1] == '{' {
			innerStart := i + 2
			innerContent, _, _, matchPos := r.parsePlaceholder(template, innerStart)
			if matchPos < 0 {
				break
			}

			// check if it's a nested start
			if strings.HasPrefix(innerContent, "range ") || strings.HasPrefix(innerContent, "if ") {
				depth++
			} else if strings.HasPrefix(innerContent, endExpr) {
				depth--
				if depth == 0 {
					// done, don't include the end marker
					endPos = matchPos + 2
					break
				}
			}
			block.WriteString(template[i : matchPos+2])
			i = matchPos + 2
		} else {
			block.WriteByte(template[i])
			i++
		}
	}

	return block.String(), endPos, nil
}

// handleIndexHelper handles the index function with proper type resolution
func (r *Renderer) handleIndexHelper(args []string) (string, error, int) {
	if len(args) < 2 {
		return "[error: index requires collection and index]", nil, 0
	}
	
	collectionArg := strings.TrimSpace(args[0])
	indexArg := strings.TrimSpace(args[1])
	
	// Resolve collection
	var collection interface{}
	if strings.HasPrefix(collectionArg, ".") {
		if val, ok := r.getVar(collectionArg); ok {
			collection = val
		} else {
			return "[error: collection not found]", nil, 0
		}
	} else {
		return "[error: collection must be a variable]", nil, 0
	}
	
	// Resolve index
	var index int
	if i, err := strconv.Atoi(indexArg); err == nil {
		index = i
	} else {
		return "[error: index must be a number]", nil, 0
	}
	
	// Try to index into the collection
	switch coll := collection.(type) {
	case []interface{}:
		if index >= 0 && index < len(coll) {
			return typeutil.ToString(coll[index]), nil, 0
		}
		return "[error: index out of range]", nil, 0
	case []string:
		if index >= 0 && index < len(coll) {
			return coll[index], nil, 0
		}
		return "[error: index out of range]", nil, 0
	default:
		if slice := typeutil.ToInterfaceSlice(collection); slice != nil {
			if index >= 0 && index < len(slice) {
				return typeutil.ToString(slice[index]), nil, 0
			}
			return "[error: index out of range]", nil, 0
		}
		return "[error: cannot index into this type]", nil, 0
	}
}

// handleIfWithConsumption processes an if block and returns output + consumed bytes
func (r *Renderer) handleIfWithConsumption(exprParts []string) (string, int) {
	condVal, ok := r.getVar("." + strings.TrimPrefix(exprParts[0], "."))
	if !ok {
		condVal = false
	}
	condition := typeutil.ToBool(condVal)
	fullBlock, consumed, err := r.extractBlockWithConsumption("if "+strings.Join(exprParts, " "), "end")
	if err != nil {
		return fmt.Sprintf("[error: %v]", err), 0
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
		return subRenderer.Render(ifBlock, nil), consumed
	} else {
		return subRenderer.Render(elseBlock, nil), consumed
	}
}

// splitParts splits a template expression into parts, respecting quoted strings and parentheses
func splitParts(s string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false
	parenDepth := 0

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"':
			inQuote = !inQuote
			current.WriteByte(s[i])
		case '(':
			if !inQuote {
				parenDepth++
			}
			current.WriteByte(s[i])
		case ')':
			if !inQuote {
				parenDepth--
			}
			current.WriteByte(s[i])
		case ' ':
			if !inQuote && parenDepth == 0 {
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
	// Look for range context in current context or any parent context
	ctx := r.context
	for ctx != nil {
		if ctx.kind == kindRange {
			// Inside a range block
			if name == "." {
				if ctx.currentRange != nil && ctx.rangeIndex < len(ctx.currentRange) {
					result := ctx.currentRange[ctx.rangeIndex]
					return result, true
				}
			} else if strings.HasPrefix(name, ".") {
				fieldName := strings.TrimPrefix(name, ".")
				// if we have a rangeMap
				if ctx.rangeMap != nil {
					if val, ok := ctx.rangeMap[fieldName]; ok {
						return val, true
					}
				}
				// or if the item is a map
				if ctx.currentRange != nil && ctx.rangeIndex < len(ctx.currentRange) {
					if item, ok := ctx.currentRange[ctx.rangeIndex].(map[string]interface{}); ok {
						if val, exists := item[fieldName]; exists {
							return val, true
						}
					}
				}
			}
			break // Found a range context, no need to look further up
		}
		ctx = ctx.parent
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

type ZeroArgHelper struct{ fn func() string }

func (h ZeroArgHelper) Execute(args []string, data map[string]interface{}) string {
	return h.fn()
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
		// Note: Variable resolution should be done before calling Execute
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

// Simple printf-style formatting helper
func printfHelper(args []string) string {
	if len(args) == 0 {
		return ""
	}
	format := args[0]
	values := make([]interface{}, len(args)-1)
	for i, arg := range args[1:] {
		values[i] = arg
	}
	return fmt.Sprintf(format, values...)
}

// Index helper to access array/slice elements or map values
func indexHelper(collection, index string) string {
	// This is a simplified implementation. In a real implementation,
	// we would need access to the actual resolved collection value.
	// For now, this is just a placeholder that demonstrates the interface.
	if i, err := strconv.Atoi(index); err == nil {
		// If index is a number, assume it's for slice/array access
		return fmt.Sprintf("[%d]", i)
	}
	// If index is not a number, assume it's for map access
	return fmt.Sprintf("[%s]", index)
}

// String filter converts input to string representation
func stringHelper(input string) string {
	return typeutil.ToString(input)
}

// Trim helper removes leading and trailing whitespace
func trimHelper(input string) string {
	return strings.TrimSpace(input)
}

func trimQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	return s
}
