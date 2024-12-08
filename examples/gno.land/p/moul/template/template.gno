package template

import (
	"strconv"
	"strings"
	"unicode"

	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
)

// Renderer is the main template processor
type Renderer struct {
	helpers  map[string]Helper
	data     map[string]interface{}
	context  []map[string]interface{}
	blocks   map[string]string
	inLoop   bool
	skipMode bool
}

// NewRenderer creates a new template renderer with default helpers
func NewRenderer() *Renderer {
	r := &Renderer{
		helpers:  make(map[string]Helper),
		data:     make(map[string]interface{}),
		context:  make([]map[string]interface{}, 0),
		blocks:   make(map[string]string),
		inLoop:   false,
		skipMode: false,
	}

	// Register markdown helpers - directly use md package functions
	r.RegisterHelper("H1", &SingleArgHelper{fn: md.H1})
	r.RegisterHelper("H2", &SingleArgHelper{fn: md.H2})
	r.RegisterHelper("H3", &SingleArgHelper{fn: md.H3})
	r.RegisterHelper("H4", &SingleArgHelper{fn: md.H4})
	r.RegisterHelper("H5", &SingleArgHelper{fn: md.H5})
	r.RegisterHelper("H6", &SingleArgHelper{fn: md.H6})
	r.RegisterHelper("Bold", &SingleArgHelper{fn: md.Bold})
	r.RegisterHelper("Italic", &SingleArgHelper{fn: md.Italic})
	r.RegisterHelper("BulletItem", &SingleArgHelper{fn: md.BulletItem})
	r.RegisterHelper("Paragraph", &SingleArgHelper{fn: md.Paragraph})
	r.RegisterHelper("Blockquote", &SingleArgHelper{fn: md.Blockquote})
	r.RegisterHelper("HorizontalRule", &SingleArgHelper{fn: func(_ string) string { return md.HorizontalRule() }})
	r.RegisterHelper("Link", &TwoArgHelper{fn: md.Link})

	// Template-specific helpers
	r.RegisterHelper("concat", NewConcatHelper())

	return r
}

// RegisterHelper adds a new helper function
func (r *Renderer) RegisterHelper(name string, helper Helper) {
	if r.helpers == nil {
		r.helpers = make(map[string]Helper)
	}
	r.helpers[name] = helper
}

func (r *Renderer) processPlaceholder(content string, data map[string]interface{}) string {
	content = strings.TrimSpace(content)

	// Handle direct variable access
	if strings.HasPrefix(content, ".") {
		return r.resolveVariable(content)
	}

	// Handle numeric literals
	if _, err := strconv.Atoi(strings.Fields(content)[0]); err == nil {
		return strings.Fields(content)[0]
	}

	parts := strings.Fields(content)
	if len(parts) == 0 {
		return ""
	}

	helperName := parts[0]
	var args string
	if len(parts) > 1 {
		args = strings.Join(parts[1:], " ")
	}

	if helper, ok := r.helpers[helperName]; ok {
		return helper.Execute(args, data)
	}

	return content
}

func (r *Renderer) handleLinkHelper(args string) string {
	// Find the position of the first and second quote pairs
	firstQuoteStart := strings.Index(args, "\"")
	if firstQuoteStart == -1 {
		return ""
	}

	firstQuoteEnd := strings.Index(args[firstQuoteStart+1:], "\"")
	if firstQuoteEnd == -1 {
		return ""
	}
	firstQuoteEnd += firstQuoteStart + 1

	secondQuoteStart := strings.Index(args[firstQuoteEnd+1:], "\"")
	if secondQuoteStart == -1 {
		return ""
	}
	secondQuoteStart += firstQuoteEnd + 1

	secondQuoteEnd := strings.Index(args[secondQuoteStart+1:], "\"")
	if secondQuoteEnd == -1 {
		return ""
	}
	secondQuoteEnd += secondQuoteStart + 1

	// Extract text and URL
	text := args[firstQuoteStart+1 : firstQuoteEnd]
	url := args[secondQuoteStart+1 : secondQuoteEnd]

	// Resolve variables in both text and URL
	text = r.resolveArguments(text)
	url = r.resolveArguments(url)

	// Format the markdown link directly
	return "[" + text + "](" + url + ")"
}

func (r *Renderer) handleRange(arg string) string {
	if !strings.HasPrefix(arg, ".") {
		return ""
	}

	collection := strings.TrimPrefix(arg, ".")
	if val, ok := r.data[collection]; ok {
		r.pushContext()
		r.inLoop = true

		// Convert the collection to []interface{}
		var items []interface{}
		switch v := val.(type) {
		case []interface{}:
			items = v
		case []string:
			items = make([]interface{}, len(v))
			for i, s := range v {
				items[i] = s
			}
		default:
			r.popContext()
			r.inLoop = false
			return ""
		}

		if len(items) > 0 {
			r.context[len(r.context)-1]["_items"] = items
			r.context[len(r.context)-1]["_index"] = 0
			r.context[len(r.context)-1]["_current"] = items[0]
		}
	}
	return ""
}

func (r *Renderer) handleIf(arg string) string {
	if !strings.HasPrefix(arg, ".") {
		return ""
	}

	condition := strings.TrimPrefix(arg, ".")
	r.pushContext()

	if val, ok := r.getVar(condition); ok {
		switch v := val.(type) {
		case bool:
			r.skipMode = !v
		case string:
			r.skipMode = v == ""
		case int:
			r.skipMode = v == 0
		case []interface{}:
			r.skipMode = len(v) == 0
		case map[string]interface{}:
			r.skipMode = len(v) == 0
		default:
			r.skipMode = false
		}
	} else {
		r.skipMode = true
	}
	return ""
}

func (r *Renderer) handlePrintf(arg string) string {
	args := splitArgs(arg)
	if len(args) < 1 {
		return ""
	}

	format := strings.Trim(args[0], "\"")
	var values []interface{}

	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, ".") {
			if val, ok := r.getVar(strings.TrimPrefix(arg, ".")); ok {
				values = append(values, val)
			} else {
				values = append(values, "<undefined>")
			}
		} else {
			values = append(values, strings.Trim(arg, "\""))
		}
	}

	return ufmt.Sprintf(format, values...)
}

func (r *Renderer) handleIndex(arg string) string {
	parts := strings.SplitN(arg, " ", 2)
	if len(parts) != 2 {
		return ""
	}

	array := strings.TrimSpace(parts[0])
	index, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return ""
	}

	if val, ok := r.getVar(strings.TrimPrefix(array, ".")); ok {
		if arr, ok := val.([]interface{}); ok {
			if index >= 0 && index < len(arr) {
				return convertToString(arr[index])
			}
		}
	}
	return ""
}

func (r *Renderer) resolveVariable(content string) string {
	if !strings.HasPrefix(content, ".") {
		return content
	}

	key := strings.TrimPrefix(content, ".")
	if val, ok := r.data[key]; ok {
		return convertToString(val)
	}
	return content
}

func (r *Renderer) handleControlFlow(helperName, args string) string {
	switch helperName {
	case "range":
		return r.handleRange(args)
	case "if":
		return r.handleIf(args)
	case "end":
		if r.inLoop {
			return r.handleRangeEnd()
		}
		r.popContext()
		return ""
	case "else":
		if r.skipMode {
			r.skipMode = false
		} else {
			r.skipMode = true
		}
		return ""
	}
	return ""
}

// isBlockElement returns true for helpers that should preserve newlines
func isBlockElement(helperName string) bool {
	blockElements := map[string]bool{
		"H1":             true,
		"H2":             true,
		"H3":             true,
		"H4":             true,
		"H5":             true,
		"H6":             true,
		"Paragraph":      true,
		"BulletItem":     true,
		"Blockquote":     true,
		"HorizontalRule": true,
	}
	return blockElements[helperName]
}

// convertToString handles conversion of interface{} to string
func convertToString(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case int:
		return ufmt.Sprintf("%d", v)
	case bool:
		return ufmt.Sprintf("%v", v)
	case map[string]interface{}:
		return ufmt.Sprintf("%v", v)
	case []interface{}:
		return ufmt.Sprintf("%v", v)
	default:
		return ufmt.Sprintf("%v", v)
	}
}

// Render processes the template with the given data
func (r *Renderer) Render(template string, data map[string]interface{}) string {
	if data != nil {
		r.data = data
	}
	r.pushContext()

	var result strings.Builder
	var lastPos int

	for {
		start := strings.Index(template[lastPos:], "{{")
		if start == -1 {
			result.WriteString(template[lastPos:])
			break
		}
		start += lastPos

		end := strings.Index(template[start:], "}}") + start
		if end == -1 {
			result.WriteString(template[lastPos:])
			break
		}

		// Extract placeholder and check for trim markers
		placeholder := template[start+2 : end]
		trimLeft := strings.HasPrefix(placeholder, "-")
		trimRight := strings.HasSuffix(placeholder, "-")

		// Handle pre-placeholder content
		if trimLeft {
			// Trim trailing whitespace from previous content
			preContent := template[lastPos:start]
			result.WriteString(strings.TrimRightFunc(preContent, unicode.IsSpace))
		} else {
			result.WriteString(template[lastPos:start])
		}

		// Process the placeholder without trim markers
		if trimLeft {
			placeholder = strings.TrimPrefix(placeholder, "-")
		}
		if trimRight {
			placeholder = strings.TrimSuffix(placeholder, "-")
		}

		output := r.processPlaceholder(strings.TrimSpace(placeholder), r.data)
		result.WriteString(output)

		// Update position and handle post-placeholder whitespace
		if trimRight {
			lastPos = end + 2
			// Skip following whitespace
			for lastPos < len(template) && unicode.IsSpace(rune(template[lastPos])) {
				lastPos++
			}
		} else {
			lastPos = end + 2
		}
	}

	r.popContext()
	return result.String()
}

// resolveNestedPlaceholders resolves placeholders inside strings.
func (r *Renderer) resolveNestedPlaceholders(content string, data map[string]interface{}) string {
	var result strings.Builder
	var lastPos int

	for {
		start := strings.Index(content[lastPos:], "{{")
		if start == -1 {
			result.WriteString(content[lastPos:])
			break
		}
		start += lastPos

		end := strings.Index(content[start:], "}}")
		if end == -1 {
			result.WriteString(content[start:])
			break
		}
		end += start + 2

		// Copy text before placeholder
		result.WriteString(content[lastPos:start])

		// Process the nested placeholder
		placeholder := strings.TrimSpace(content[start+2 : end-2])
		resolved := r.processPlaceholder(placeholder, data)
		result.WriteString(resolved)

		lastPos = end
	}

	return result.String()
}

// processArgument processes arguments and resolves placeholders if present.
func (r *Renderer) processArgument(arg string, data map[string]interface{}) string {
	if strings.Contains(arg, "{{") && strings.Contains(arg, "}}") {
		return r.resolveNestedPlaceholders(arg, data)
	}

	if len(arg) > 1 && arg[0] == '"' && arg[len(arg)-1] == '"' {
		return arg[1 : len(arg)-1] // Strip quotes
	}

	return arg
}

// parseArgs parses arguments for helper calls.
func (r *Renderer) parseArgs(args []string, data map[string]interface{}) []interface{} {
	var parsedArgs []interface{}
	for _, arg := range args {
		parsed := r.processArgument(arg, data)
		parsedArgs = append(parsedArgs, parsed)
	}
	return parsedArgs
}

// Context management methods
func (r *Renderer) pushContext() {
	r.context = append(r.context, make(map[string]interface{}))
}

func (r *Renderer) popContext() {
	if len(r.context) > 0 {
		r.context = r.context[:len(r.context)-1]
	}
}

func (r *Renderer) setVar(name string, value interface{}) {
	if len(r.context) > 0 {
		r.context[len(r.context)-1][name] = value
	}
}

func (r *Renderer) getVar(name string) (interface{}, bool) {
	// Check current context first
	if len(r.context) > 0 {
		currentContext := r.context[len(r.context)-1]
		if val, ok := currentContext[name]; ok {
			return val, true
		}
	}
	// Fall back to global data
	val, ok := r.data[name]
	return val, ok
}

// Additional methods needed:
func (r *Renderer) pushBlock(name, content string) {
	r.blocks[name] = content
}

func (r *Renderer) getBlock(name string) (string, bool) {
	content, ok := r.blocks[name]
	return content, ok
}

func (r *Renderer) enterLoop() {
	r.inLoop = true
}

func (r *Renderer) exitLoop() {
	r.inLoop = false
}

func (r *Renderer) setSkipMode(skip bool) {
	r.skipMode = skip
}

func (r *Renderer) isSkipping() bool {
	return r.skipMode
}

func (r *Renderer) resolveVariablesInPrintf(format string) string {
	args := splitArgs(format)
	if len(args) < 1 {
		return format
	}

	var resolvedArgs []string
	resolvedArgs = append(resolvedArgs, args[0]) // Keep format string

	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, ".") {
			resolved := r.resolveVariable(arg)
			resolvedArgs = append(resolvedArgs, resolved)
		} else {
			resolvedArgs = append(resolvedArgs, arg)
		}
	}

	return strings.Join(resolvedArgs, " ")
}

func (r *Renderer) handleRangeEnd() string {
	currentCtx := r.context[len(r.context)-1]
	if items, ok := currentCtx["_items"].([]interface{}); ok {
		if idx, ok := currentCtx["_index"].(int); ok {
			idx++
			if idx < len(items) {
				currentCtx["_index"] = idx
				currentCtx["_current"] = items[idx]

				// Handle next item based on its type
				switch item := items[idx].(type) {
				case string:
					// Pass both the argument and the data context
					return r.helpers["BulletItem"].Execute(item, r.data)
				case map[string]interface{}:
					// Clear previous item's keys
					for k := range currentCtx {
						if k != "_items" && k != "_index" && k != "_current" {
							delete(currentCtx, k)
						}
					}
					// Set new item's keys
					for k, v := range item {
						currentCtx[k] = v
					}
				}
				return ""
			}
		}
	}

	r.inLoop = false
	r.popContext()
	return ""
}

func (r *Renderer) resolveArguments(args string) string {
	// Check for nested placeholders
	if strings.Contains(args, "{{") && strings.Contains(args, "}}") {
		return r.resolveNestedPlaceholders(args, r.data)
	}

	// Handle direct variable references
	if strings.HasPrefix(args, ".") {
		if val, ok := r.data[strings.TrimPrefix(args, ".")]; ok {
			return convertToString(val)
		}
	}

	// Remove quotes if present
	if strings.HasPrefix(args, "\"") && strings.HasSuffix(args, "\"") {
		return args[1 : len(args)-1]
	}

	return args
}

func (r *Renderer) resolveLoopVariable(content string) string {
	if !strings.HasPrefix(content, ".") {
		return content
	}

	// If it's just ".", return the current item
	if content == "." {
		currentCtx := r.context[len(r.context)-1]
		if current, ok := currentCtx["_current"]; ok {
			return convertToString(current)
		}
		return "<undefined>"
	}

	key := strings.TrimPrefix(content, ".")
	currentCtx := r.context[len(r.context)-1]

	// Try current item if it's a map
	if current, ok := currentCtx["_current"]; ok {
		if m, ok := current.(map[string]interface{}); ok {
			if val, ok := m[key]; ok {
				return convertToString(val)
			}
		}
	}

	// Try current context
	if val, ok := currentCtx[key]; ok {
		return convertToString(val)
	}

	// Fall back to global data
	if val, ok := r.data[key]; ok {
		return convertToString(val)
	}

	return key // Return the key itself instead of undefined message
}

func (r *Renderer) handlePrintfInLoop(args string) string {
	format, vars := r.parsePrintfArgs(args)
	resolvedVars := make([]interface{}, len(vars))

	for i, v := range vars {
		if strings.HasPrefix(v, ".") {
			resolvedVars[i] = r.resolveLoopVariable(v)
		} else {
			resolvedVars[i] = v
		}
	}

	return ufmt.Sprintf(format, resolvedVars...)
}

func (r *Renderer) parsePrintfArgs(args string) (string, []string) {
	// Extract format string (in quotes) and variables
	var format string
	var vars []string

	if strings.HasPrefix(args, "\"") {
		endQuote := strings.Index(args[1:], "\"") + 1
		if endQuote > 0 {
			format = args[1:endQuote]
			vars = strings.Fields(args[endQuote+1:])
		}
	}

	return format, vars
}

func (r *Renderer) handleBulletItem(content string) string {
	if strings.HasPrefix(content, "-") {
		return content
	}
	return "- " + content
}

// splitArgs splits a string into arguments, preserving quoted strings
func splitArgs(s string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"':
			inQuotes = !inQuotes
			current.WriteByte(s[i])
		case ' ':
			if inQuotes {
				current.WriteByte(s[i])
			} else if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(s[i])
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}
