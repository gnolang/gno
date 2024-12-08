package template

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"gno.land/p/moul/md"
	"gno.land/p/moul/typeutil"
)

// Add at the top of the file, after package declaration
var debug bool = false

func log(args ...interface{}) {
	if !debug {
		return
	}
	// Convert args to strings and join them
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

// Context management
type context struct {
	parent       *context
	currentRange []interface{}
	rangeMap     map[string]interface{}
	rangeIndex   int
	kind         int
}

// Renderer is the main template processor
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

	// Register other helpers (in separate category)
	r.helpers["concat"] = VarArgHelper{fn: concatHelper}

	return r
}

func (r *Renderer) processPlaceholder(content string) (string, error) {
	parts := splitParts(content)
	if len(parts) == 0 {
		return "", nil
	}

	// First check if it's just a variable reference (starts with dot)
	if strings.HasPrefix(parts[0], ".") {
		val, ok := r.getVar(parts[0])
		if ok {
			return typeutil.ToString(val), nil
		}
		// Return empty string if variable not found instead of the variable name
		return "", nil
	}

	switch parts[0] {
	case "range":
		log("=== Processing range placeholder ===")
		log("Content:", content)
		if len(parts) < 2 {
			return "", errors.New("range requires a variable")
		}
		val, ok := r.getVar(parts[1])
		if !ok {
			log("Variable not found:", parts[1])
			return "", nil
		}
		log("Variable value:", val)

		rangeContent := r.extractRangeContent(strings.Join(parts[1:], " "))
		log("Extracted range content:", rangeContent)
		if rangeContent == "" {
			return "", fmt.Errorf("could not find range content")
		}

		var result strings.Builder
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

		for i, item := range rangeData {
			newContext := &context{
				parent:       r.context,
				currentRange: rangeData,
				rangeIndex:   i,
				kind:         kindRange,
				rangeMap:     nil,
			}

			if m, ok := item.(map[string]interface{}); ok {
				newContext.rangeMap = m
			}

			subRenderer := &Renderer{
				helpers: r.helpers,
				data:    r.data,
				context: newContext,
			}

			processed := subRenderer.Render(rangeContent, nil)
			result.WriteString(processed)
		}

		log("Final range result:", result.String())
		return result.String(), nil

	case "end":
		return "", nil

	case "if":
		if len(parts) < 2 {
			return "", fmt.Errorf("if requires a condition")
		}
		val, ok := r.getVar(parts[1])
		if !ok {
			val = false
		}
		condition := typeutil.ToBool(val)

		ifContent := r.extractIfContent(strings.Join(parts[1:], " "))
		if ifContent == "" {
			return "", nil
		}

		if condition {
			newContext := &context{
				parent: r.context,
				kind:   kindIf,
			}

			subRenderer := &Renderer{
				helpers:     r.helpers,
				data:        r.data,
				context:     newContext,
				lastContent: ifContent,
			}
			return subRenderer.Render(ifContent, nil), nil
		}
		return "", nil

	default:
		// Check for helpers
		if helper, ok := r.helpers[parts[0]]; ok {
			return helper.Execute(parts[1:], r.data), nil
		}

		// If not a helper and not a special command, treat as plain text
		return strings.Join(parts, " "), nil
	}
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

func trimMarkers(content string) (string, bool, bool) {
	content = strings.TrimSpace(content)
	trimLeft := false
	trimRight := false

	if strings.HasPrefix(content, "-") {
		trimLeft = true
		content = strings.TrimPrefix(content, "-")
		content = strings.TrimLeftFunc(content, unicode.IsSpace)
	}
	if strings.HasSuffix(content, "-") {
		trimRight = true
		content = strings.TrimSuffix(content, "-")
		content = strings.TrimRightFunc(content, unicode.IsSpace)
	}
	return strings.TrimSpace(content), trimLeft, trimRight
}

// Render processes the template with the given data
func (r *Renderer) Render(template string, data map[string]interface{}) string {
	if data != nil {
		r.data = data
	}

	var result strings.Builder
	lastPos := 0
	r.lastContent = template

	for i := 0; i < len(template); {
		if i+1 < len(template) && template[i] == '{' && template[i+1] == '{' {
			end := i + 2
			depth := 1
			for end < len(template)-1 {
				if template[end] == '}' && template[end+1] == '}' {
					depth--
					if depth == 0 {
						break
					}
				} else if template[end] == '{' && template[end+1] == '{' {
					depth++
				}
				end++
			}

			if end < len(template)-1 {
				content := template[i+2 : end]
				content, trimLeft, trimRight := trimMarkers(content)

				if trimLeft {
					result.WriteString(strings.TrimRightFunc(template[lastPos:i], unicode.IsSpace))
				} else {
					result.WriteString(template[lastPos:i])
				}

				output, err := r.processPlaceholder(content)
				if err != nil {
					log("Error processing placeholder:", err)
					return fmt.Sprintf("[error: %v]", err)
				}
				result.WriteString(output)

				i = end + 2
				lastPos = i

				if trimRight {
					for lastPos < len(template) && unicode.IsSpace(rune(template[lastPos])) {
						lastPos++
						i++
					}
				}
				continue
			}
		}
		i++
	}

	if lastPos < len(template) {
		result.WriteString(template[lastPos:])
	}

	return result.String()
}

// Context management methods
func (r *Renderer) pushContext(data interface{}, kind int) {
	nc := &context{
		parent: r.context,
		kind:   kind,
	}

	switch v := data.(type) {
	case []interface{}:
		nc.currentRange = v
		nc.rangeIndex = 0
	case map[string]interface{}:
		nc.rangeMap = v
	case bool:
		nc.currentRange = []interface{}{v}
	}

	r.context = nc
}

func (r *Renderer) popContext() {
	if r.context != nil {
		r.context = r.context.parent
	}
}

// getVar retrieves a variable from the current context or data
func (r *Renderer) getVar(name string) (interface{}, bool) {
	if r.context != nil && r.context.kind == kindRange {
		if name == "." {
			if r.context.currentRange != nil && r.context.rangeIndex < len(r.context.currentRange) {
				return r.context.currentRange[r.context.rangeIndex], true
			}
		} else if strings.HasPrefix(name, ".") {
			name = strings.TrimPrefix(name, ".")
			// First check rangeMap if available
			if r.context.rangeMap != nil {
				if val, exists := r.context.rangeMap[name]; exists {
					return val, true
				}
			}
			// Then check current range item if it's a map
			if r.context.currentRange != nil && r.context.rangeIndex < len(r.context.currentRange) {
				if item, ok := r.context.currentRange[r.context.rangeIndex].(map[string]interface{}); ok {
					if val, exists := item[name]; exists {
						return val, true
					}
				}
			}
		}
	}

	// If not found in context, look in data
	return r.lookupVar(strings.TrimPrefix(name, "."))
}

// lookupVar looks up a variable in a map by name, supporting dot notation
func (r *Renderer) lookupVar(name string) (interface{}, bool) {
	// Split the name into parts for nested lookups
	parts := strings.Split(name, ".")

	current := r.data
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - return the value
			val, ok := current[part]
			return val, ok
		}

		// Not the last part - must be a map
		next, ok := current[part]
		if !ok {
			return nil, false
		}

		// Convert to map for next iteration
		nextMap, ok := next.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current = nextMap
	}

	return nil, false
}

// Helper interface definition
type Helper interface {
	Execute(args []string, data map[string]interface{}) string
}

// SingleArgHelper implementation
type SingleArgHelper struct {
	fn func(string) string
}

func (h SingleArgHelper) Execute(args []string, data map[string]interface{}) string {
	if len(args) < 1 {
		return "[error: missing argument]"
	}

	// Join all args for quoted strings that might contain spaces
	arg := strings.Join(args, " ")

	// Handle quoted strings properly
	if strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"") {
		// Remove outer quotes and keep inner quotes as-is
		arg = arg[1 : len(arg)-1]
	}

	return h.fn(arg)
}

// VarArgHelper implementation
type VarArgHelper struct {
	fn func([]string) string
}

func (h VarArgHelper) Execute(args []string, data map[string]interface{}) string {
	processedArgs := make([]string, len(args))
	for i, arg := range args {
		arg = strings.TrimSpace(arg)
		if strings.HasPrefix(arg, ".") {
			if val, ok := data[strings.TrimPrefix(arg, ".")]; ok {
				processedArgs[i] = typeutil.ToString(val)
				continue
			}
		} else if strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"") {
			processedArgs[i] = arg[1 : len(arg)-1]
			continue
		}
		processedArgs[i] = arg
	}
	return h.fn(processedArgs)
}

// TwoArgHelper implementation
type TwoArgHelper struct {
	fn func(string, string) string
}

func (h TwoArgHelper) Execute(args []string, data map[string]interface{}) string {
	if len(args) < 2 {
		return "[error: requires two arguments]"
	}

	// Process first argument
	arg1 := args[0]
	if strings.HasPrefix(arg1, "\"") && strings.HasSuffix(arg1, "\"") {
		arg1 = arg1[1 : len(arg1)-1]
	}

	// Process second argument
	arg2 := args[1]
	if strings.HasPrefix(arg2, "\"") && strings.HasSuffix(arg2, "\"") {
		arg2 = arg2[1 : len(arg2)-1]
	}

	return h.fn(arg1, arg2)
}

// TwoArgStringBoolHelper implementation
type TwoArgStringBoolHelper struct {
	fn func(string, bool) string
}

func (h TwoArgStringBoolHelper) Execute(args []string, data map[string]interface{}) string {
	if len(args) < 2 {
		return "[error: requires two arguments]"
	}

	// Process first argument as a string
	arg1 := args[0]
	if strings.HasPrefix(arg1, "\"") && strings.HasSuffix(arg1, "\"") {
		arg1 = arg1[1 : len(arg1)-1]
	}

	// Process second argument as a boolean
	arg2 := strings.ToLower(args[1]) == "true"

	return h.fn(arg1, arg2)
}

// ThreeArgHelper implementation
type ThreeArgHelper struct {
	fn func(string, string, string) string
}

func (h ThreeArgHelper) Execute(args []string, data map[string]interface{}) string {
	if len(args) < 3 {
		return "[error: requires three arguments]"
	}

	// Process first argument
	arg1 := args[0]
	if strings.HasPrefix(arg1, "\"") && strings.HasSuffix(arg1, "\"") {
		arg1 = arg1[1 : len(arg1)-1]
	}

	// Process second argument
	arg2 := args[1]
	if strings.HasPrefix(arg2, "\"") && strings.HasSuffix(arg2, "\"") {
		arg2 = arg2[1 : len(arg2)-1]
	}

	// Process third argument
	arg3 := args[2]
	if strings.HasPrefix(arg3, "\"") && strings.HasSuffix(arg3, "\"") {
		arg3 = arg3[1 : len(arg3)-1]
	}

	return h.fn(arg1, arg2, arg3)
}

// Helper function for string concatenation
func concatHelper(args []string) string {
	return strings.Join(args, "")
}

// New helper function to extract range content
func (r *Renderer) extractRangeContent(rangeExpr string) string {
	log("=== extractRangeContent Debug ===")
	log("Range Expression:", rangeExpr)
	log("Last Content:", r.lastContent)

	// Look for the range start marker
	marker := "{{ range " + strings.TrimSpace(rangeExpr) + " }}"
	start := strings.Index(r.lastContent, marker)
	if start < 0 {
		// Try without spaces as fallback
		marker = "{{range " + strings.TrimSpace(rangeExpr) + "}}"
		start = strings.Index(r.lastContent, marker)
		if start < 0 {
			log("Could not find range start marker")
			return ""
		}
	}
	log("Found range start marker at position:", start)
	log("Marker:", marker)

	contentStart := start + len(marker)
	depth := 1
	endPos := -1

	log("Starting content scan from position:", contentStart)
	for i := contentStart; i < len(r.lastContent)-5; i++ {
		if i+2 <= len(r.lastContent) && r.lastContent[i:i+2] == "{{" {
			// Skip any whitespace after {{
			j := i + 2
			for j < len(r.lastContent) && unicode.IsSpace(rune(r.lastContent[j])) {
				j++
			}

			// Check for range or end keywords
			rest := r.lastContent[j:]
			if strings.HasPrefix(rest, "range") {
				log("Found nested range at position:", i)
				depth++
				log("Depth increased to:", depth)
			} else if strings.HasPrefix(rest, "end") {
				log("Found end tag at position:", i)
				depth--
				log("Depth decreased to:", depth)
				if depth == 0 {
					endPos = i
					log("Found matching end tag at position:", endPos)
					break
				}
			}
		}
	}

	if endPos == -1 {
		log("No matching end tag found")
		return ""
	}

	content := r.lastContent[contentStart:endPos]
	log("Extracted content:", content)
	return content
}

// Add a similar method for if blocks
func (r *Renderer) extractIfContent(ifExpr string) string {
	marker := "{{ if " + strings.TrimSpace(ifExpr) + " }}"
	start := strings.Index(r.lastContent, marker)
	if start < 0 {
		marker = "{{if " + strings.TrimSpace(ifExpr) + "}}"
		start = strings.Index(r.lastContent, marker)
		if start < 0 {
			return ""
		}
	}

	contentStart := start + len(marker)
	depth := 1
	endPos := -1
	elsePos := -1

	for i := contentStart; i < len(r.lastContent)-5; i++ {
		if i+2 <= len(r.lastContent) && r.lastContent[i:i+2] == "{{" {
			rest := r.lastContent[i+2:]
			if strings.HasPrefix(rest, " if ") || strings.HasPrefix(rest, "if ") {
				depth++
			} else if strings.HasPrefix(rest, " end ") || strings.HasPrefix(rest, "end ") {
				depth--
				if depth == 0 {
					endPos = i
					break
				}
			} else if depth == 1 && (strings.HasPrefix(rest, " else ") || strings.HasPrefix(rest, "else ")) {
				elsePos = i
			}
		}
	}

	if endPos == -1 {
		return ""
	}

	if elsePos != -1 {
		return r.lastContent[contentStart:elsePos]
	}
	return r.lastContent[contentStart:endPos]
}
