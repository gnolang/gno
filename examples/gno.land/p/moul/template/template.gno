package template

import (
	"fmt"
	"strings"
	"unicode"

	"gno.land/p/moul/md"
	"gno.land/p/moul/typeutil"
)

var debug = false // Enable logging

// Context kinds
const (
	kindNone = iota
	kindRange
	kindIf
)

func log(args ...interface{}) {
	if debug {
		println(args...)
	}
}

// Context management
type context struct {
	parent       *context
	currentRange []interface{}
	rangeMap     map[string]interface{}
	rangeIndex   int
	kind         int
	next         int
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
	r.helpers["Bold"] = SingleArgHelper{fn: md.Bold}
	r.helpers["H1"] = SingleArgHelper{fn: md.H1}
	r.helpers["BulletItem"] = SingleArgHelper{fn: md.BulletItem}
	r.helpers["concat"] = VarArgHelper{fn: concatHelper}

	return r
}

func (r *Renderer) processPlaceholder(content string) (string, error) {
	parts := splitParts(content)
	log("processPlaceholder - content:", content, "parts:", parts)

	// Handle special commands
	if len(parts) > 0 {
		switch parts[0] {
		case "range":
			if len(parts) < 2 {
				return "", fmt.Errorf("range requires a variable")
			}
			val, ok := r.getVar(parts[1])
			if !ok {
				return "", nil
			}
			log("  range: got value:", val, ok)

			// Convert any slice type to []interface{}
			var rangeData []interface{}
			switch v := val.(type) {
			case []interface{}:
				rangeData = v
			case []int:
				rangeData = typeutil.ToInterfaceSlice(v)
			case []string:
				rangeData = typeutil.ToInterfaceSlice(v)
			case map[string]interface{}:
				r.pushContext(v, kindRange)
				return "", nil
			case []map[string]interface{}:
				return "", fmt.Errorf("unsupported type: []map[string]interface{}")
			default:
				if convertedMap, err := typeutil.ToMapStringInterface(v); err == nil {
					r.pushContext(convertedMap, kindRange)
					return "", nil
				} else {
					return "", fmt.Errorf("cannot range over %T", val)
				}
			}

			r.pushContext(rangeData, kindRange)
			r.context.next = 0
			return "", nil

		case "end":
			if r.context == nil {
				return "", fmt.Errorf("unexpected end")
			}
			r.popContext()
			return "", nil

		case ".":
			// Handle dot operator in range context
			if r.context != nil && r.context.kind == kindRange {
				if r.context.rangeIndex < len(r.context.currentRange) {
					return typeutil.ToString(r.context.currentRange[r.context.rangeIndex]), nil
				}
			}
			return "", nil
		}
	}

	// Handle variables and helpers
	if strings.HasPrefix(parts[0], ".") {
		val, ok := r.getVar(parts[0])
		if !ok {
			return parts[0], nil // Return raw text if variable not found
		}
		return typeutil.ToString(val), nil
	}

	// Handle helpers
	if helper, ok := r.helpers[parts[0]]; ok {
		return helper.Execute(parts[1:], r.data), nil
	}

	return strings.Join(parts, " "), nil
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

// Fix the trim marker detection functions
func hasTrimLeft(content string) bool {
	content = strings.TrimSpace(content)
	return strings.HasPrefix(content, "-")
}

func hasTrimRight(content string) bool {
	content = strings.TrimSpace(content)
	return strings.HasSuffix(content, "-")
}

func trimMarkers(content string) string {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "-") {
		content = strings.TrimPrefix(content, "-")
		content = strings.TrimLeftFunc(content, unicode.IsSpace)
	}
	if strings.HasSuffix(content, "-") {
		content = strings.TrimSuffix(content, "-")
		content = strings.TrimRightFunc(content, unicode.IsSpace)
	}
	return strings.TrimSpace(content)
}

// Render processes the template with the given data
func (r *Renderer) Render(template string, data map[string]interface{}) string {
	r.data = data
	var result strings.Builder
	lastPos := 0
	var lastTrimRight bool
	var nextTrimLeft bool

	log("Render: starting template processing")

	for i := 0; i < len(template); i++ {
		if template[i] == '{' && i+1 < len(template) && template[i+1] == '{' {
			// Add text before placeholder
			if i > lastPos {
				text := template[lastPos:i]
				log("Text before placeholder:", fmt.Sprintf("%q", text))
				log("lastTrimRight:", lastTrimRight, "nextTrimLeft:", nextTrimLeft)

				// Extract and process content
				end := i + 2
				for end < len(template)-1 {
					if template[end] == '}' && template[end+1] == '}' {
						break
					}
					end++
				}

				if end < len(template)-1 {
					content := template[i+2 : end]
					trimLeft := hasTrimLeft(content)
					nextTrimLeft = trimLeft
				}

				// Apply trims to the text between placeholders
				if lastTrimRight && nextTrimLeft {
					// Keep only non-whitespace characters
					var nonWS strings.Builder
					for _, ch := range text {
						if !unicode.IsSpace(ch) {
							nonWS.WriteRune(ch)
						}
					}
					text = nonWS.String()
					log("Both trims active, keeping non-whitespace:", fmt.Sprintf("%q", text))
				} else {
					if lastTrimRight {
						text = strings.TrimLeftFunc(text, unicode.IsSpace)
					}
					if nextTrimLeft {
						text = strings.TrimRightFunc(text, unicode.IsSpace)
					}
				}

				log("After trim text:", fmt.Sprintf("%q", text))
				result.WriteString(text)
			}

			// Find end of placeholder
			end := i + 2
			for end < len(template)-1 {
				if template[end] == '}' && template[end+1] == '}' {
					break
				}
				end++
			}

			if end < len(template)-1 {
				// Extract and process content
				content := template[i+2 : end]
				log("Found placeholder:", fmt.Sprintf("%q", content))

				// Check trim markers before processing content
				trimLeft := hasTrimLeft(content)
				trimRight := hasTrimRight(content)
				log("Trim markers - left:", trimLeft, "right:", trimRight)

				// Update trim flags for next iteration
				nextTrimLeft = trimLeft
				lastTrimRight = trimRight

				// Process the placeholder
				content = trimMarkers(content)
				output, err := r.processPlaceholder(content)
				if err != nil {
					return fmt.Sprintf("[error: %v]", err)
				}

				log("Placeholder output:", fmt.Sprintf("%q", output))
				result.WriteString(output)

				i = end + 1
				lastPos = i + 1
			}
		}
	}

	// Add remaining text
	if lastPos < len(template) {
		text := template[lastPos:]
		log("Remaining text:", fmt.Sprintf("%q", text))
		if lastTrimRight {
			text = strings.TrimLeftFunc(text, unicode.IsSpace)
		}
		log("Final remaining text:", fmt.Sprintf("%q", text))
		result.WriteString(text)
	}

	final := result.String()
	log("Final result:", fmt.Sprintf("%q", final))
	return final
}

// Context management methods
func (r *Renderer) pushContext(data interface{}, kind int) {
	nc := &context{
		parent: r.context,
		kind:   kind,
	}

	// Handle different range data types
	if kind == kindRange {
		switch v := data.(type) {
		case []interface{}:
			nc.currentRange = v
		case map[string]interface{}:
			nc.rangeMap = v
		}
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
	log("getVar: context:", r.context)

	// Handle range context first
	if r.context != nil && r.context.kind == kindRange {
		if strings.HasPrefix(name, ".") {
			if r.context.rangeIndex < len(r.context.currentRange) {
				currentItem := r.context.currentRange[r.context.rangeIndex]
				name = strings.TrimPrefix(name, ".")
				if name == "" {
					return currentItem, true
				}
				if m, ok := currentItem.(map[string]interface{}); ok {
					if val, exists := m[name]; exists {
						return val, true
					}
				}
			}
		}
	}

	// Look up in data
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
		// Remove outer quotes and unescape inner quotes
		arg = arg[1 : len(arg)-1]
		arg = strings.ReplaceAll(arg, "\\\"", "\"")
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

// Helper function for string concatenation
func concatHelper(args []string) string {
	return strings.Join(args, "")
}
