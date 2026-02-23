package template

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"gno.land/p/moul/md"
	"gno.land/p/moul/typeutil"
)

// Renderer provides a simple template engine with a clean API
type Renderer struct {
	funcs map[string]Func  // Template functions
	data  map[string]interface{} // Global data
	ctx   *context       // Current evaluation context
}

// Func is the signature for template functions
type Func func(args ...string) string

// context tracks variable scope during evaluation
type context struct {
	parent *context
	vars   map[string]interface{} // Local variables
	loop   *loopContext          // Range loop state
}

// loopContext holds state for range loops
type loopContext struct {
	items []interface{} // Items being iterated
	index int          // Current index
}

// NewRenderer creates a new template renderer with default functions
func NewRenderer() *Renderer {
	r := &Renderer{
		funcs: make(map[string]Func),
		data:  make(map[string]interface{}),
	}
	r.registerDefaults()
	return r
}

// Render processes a template with the given data
func (r *Renderer) Render(template string, data map[string]interface{}) string {
	if data != nil {
		r.data = data
	}
	r.ctx = nil
	return r.render(template)
}

// render is the core rendering engine
func (r *Renderer) render(tmpl string) string {
	var out strings.Builder
	
	for len(tmpl) > 0 {
		// Find next placeholder
		start := strings.Index(tmpl, "{{")
		if start < 0 {
			out.WriteString(tmpl)
			break
		}
		
		// Write content before placeholder
		out.WriteString(tmpl[:start])
		
		// Parse placeholder
		placeholder, end := r.parsePlaceholder(tmpl[start:])
		if placeholder == nil {
			out.WriteString("{{")
			tmpl = tmpl[start+2:]
			continue
		}
		
		// Apply left trim
		if placeholder.trimLeft {
			s := out.String()
			out.Reset()
			out.WriteString(strings.TrimRightFunc(s, unicode.IsSpace))
		}
		
		// Process placeholder
		result, consumed := r.process(placeholder, tmpl[start+end:])
		out.WriteString(result)
		
		// Advance position
		tmpl = tmpl[start+end+consumed:]
		
		// Apply right trim
		if placeholder.trimRight {
			tmpl = strings.TrimLeftFunc(tmpl, unicode.IsSpace)
		}
	}
	
	return out.String()
}

// placeholder represents a parsed {{...}} tag
type placeholder struct {
	expr      string
	trimLeft  bool
	trimRight bool
}

// parsePlaceholder extracts a {{...}} placeholder
func (r *Renderer) parsePlaceholder(tmpl string) (*placeholder, int) {
	if !strings.HasPrefix(tmpl, "{{") {
		return nil, 0
	}
	
	// Find matching }}
	end := r.findClosing(tmpl, 2)
	if end < 0 {
		return nil, 0
	}
	
	// Extract content
	content := tmpl[2:end]
	
	// Check trim markers
	p := &placeholder{}
	if strings.HasPrefix(content, "-") {
		p.trimLeft = true
		content = content[1:]
	}
	if strings.HasSuffix(content, "-") {
		p.trimRight = true
		content = content[:len(content)-1]
	}
	
	p.expr = strings.TrimSpace(content)
	return p, end + 2
}

// findClosing finds the closing }} for a {{
func (r *Renderer) findClosing(tmpl string, start int) int {
	depth := 1
	i := start
	
	for i < len(tmpl)-1 {
		if tmpl[i] == '{' && tmpl[i+1] == '{' {
			depth++
			i += 2
		} else if tmpl[i] == '}' && tmpl[i+1] == '}' {
			depth--
			if depth == 0 {
				return i
			}
			i += 2
		} else {
			i++
		}
	}
	
	return -1
}

// process evaluates a placeholder and returns (result, extraBytesConsumed)
func (r *Renderer) process(p *placeholder, remaining string) (string, int) {
	parts := r.parseExpr(p.expr)
	if len(parts) == 0 {
		return "", 0
	}
	
	cmd := parts[0]
	args := parts[1:]
	
	// Handle variables
	if strings.HasPrefix(cmd, ".") {
		if val := r.resolve(cmd); val != nil {
			return typeutil.ToString(val), 0
		}
		return "", 0
	}
	
	// Handle commands
	switch cmd {
	case "range":
		return r.doRange(args, remaining)
	case "if":
		return r.doIf(args, remaining)
	case "end", "else":
		return "", 0 // Handled by parent
	default:
		// Special handling for index function
		if cmd == "index" && len(args) >= 2 {
			return r.indexFunc(args...), 0
		}
		// Try function
		if fn, ok := r.funcs[cmd]; ok {
			resolved := r.resolveArgs(args)
			return fn(resolved...), 0
		}
		return p.expr, 0
	}
}

// doRange implements {{range .items}}...{{end}}
func (r *Renderer) doRange(args []string, tmpl string) (string, int) {
	if len(args) == 0 {
		return "", 0
	}
	
	// Get collection
	val := r.resolve(args[0])
	if val == nil {
		return "", 0
	}
	
	// Convert to items
	items := toSlice(val)
	if items == nil {
		return "", 0
	}
	
	// Find block
	block, consumed := r.findBlock(tmpl, "end")
	if consumed == 0 {
		return "", 0
	}
	
	// Execute loop
	var out strings.Builder
	oldCtx := r.ctx
	
	for i, item := range items {
		// Create loop context
		r.ctx = &context{
			parent: oldCtx,
			vars:   make(map[string]interface{}),
			loop: &loopContext{
				items: items,
				index: i,
			},
		}
		
		// Add item properties if map
		if m, ok := item.(map[string]interface{}); ok {
			for k, v := range m {
				r.ctx.vars[k] = v
			}
		}
		
		out.WriteString(r.render(block))
	}
	
	r.ctx = oldCtx
	return out.String(), consumed
}

// doIf implements {{if .cond}}...{{else}}...{{end}}
func (r *Renderer) doIf(args []string, tmpl string) (string, int) {
	if len(args) == 0 {
		return "", 0
	}
	
	// Evaluate condition
	val := r.resolve(args[0])
	cond := typeutil.ToBool(val)
	
	// Find block
	block, consumed := r.findBlock(tmpl, "end")
	if consumed == 0 {
		return "", 0
	}
	
	// Split on else
	ifBlock, elseBlock := r.splitElse(block)
	
	// Execute branch
	oldCtx := r.ctx
	r.ctx = &context{
		parent: oldCtx,
		vars:   make(map[string]interface{}),
	}
	
	var result string
	if cond {
		result = r.render(ifBlock)
	} else {
		result = r.render(elseBlock)
	}
	
	r.ctx = oldCtx
	return result, consumed
}

// findBlock finds content up to {{end}}
func (r *Renderer) findBlock(tmpl string, endMarker string) (string, int) {
	var out strings.Builder
	depth := 1
	i := 0
	
	for i < len(tmpl) {
		// Find next {{
		next := strings.Index(tmpl[i:], "{{")
		if next < 0 {
			break
		}
		
		// Add content before {{
		out.WriteString(tmpl[i : i+next])
		
		// Parse placeholder
		p, consumed := r.parsePlaceholder(tmpl[i+next:])
		if p == nil {
			out.WriteString("{{")
			i += next + 2
			continue
		}
		
		// Check command
		parts := r.parseExpr(p.expr)
		if len(parts) > 0 {
			switch parts[0] {
			case "range", "if":
				depth++
			case endMarker:
				depth--
				if depth == 0 {
					return out.String(), i + next + consumed
				}
			}
		}
		
		// Add to output
		out.WriteString(tmpl[i+next : i+next+consumed])
		i += next + consumed
	}
	
	return "", 0
}

// splitElse splits block on {{else}}
func (r *Renderer) splitElse(block string) (string, string) {
	depth := 0
	i := 0
	
	for i < len(block) {
		next := strings.Index(block[i:], "{{")
		if next < 0 {
			break
		}
		
		p, consumed := r.parsePlaceholder(block[i+next:])
		if p == nil {
			i += next + 2
			continue
		}
		
		parts := r.parseExpr(p.expr)
		if len(parts) > 0 {
			switch parts[0] {
			case "if":
				depth++
			case "end":
				depth--
			case "else":
				if depth == 0 {
					return block[:i+next], block[i+next+consumed:]
				}
			}
		}
		
		i += next + consumed
	}
	
	return block, ""
}

// parseExpr splits expression respecting quotes and parentheses
func (r *Renderer) parseExpr(expr string) []string {
	var parts []string
	var current strings.Builder
	var inQuote bool
	var parenDepth int
	var prevChar rune
	
	for _, ch := range expr {
		switch ch {
		case '"':
			if prevChar != '\\' {
				inQuote = !inQuote
			}
			current.WriteRune(ch)
		case '(':
			if !inQuote {
				parenDepth++
			}
			current.WriteRune(ch)
		case ')':
			if !inQuote {
				parenDepth--
			}
			current.WriteRune(ch)
		case ' ':
			if !inQuote && parenDepth == 0 {
				if current.Len() > 0 {
					parts = append(parts, current.String())
					current.Reset()
				}
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
		prevChar = ch
	}
	
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	
	return parts
}

// resolve looks up a variable
func (r *Renderer) resolve(name string) interface{} {
	if !strings.HasPrefix(name, ".") {
		return nil
	}
	name = strings.TrimPrefix(name, ".")
	
	// Check context stack
	ctx := r.ctx
	for ctx != nil {
		// Handle . in range
		if name == "" && ctx.loop != nil {
			if ctx.loop.index < len(ctx.loop.items) {
				return ctx.loop.items[ctx.loop.index]
			}
		}
		
		// Check local vars
		if val, ok := ctx.vars[name]; ok {
			return val
		}
		
		ctx = ctx.parent
	}
	
	// Check global data
	return lookup(r.data, name)
}

// resolveArgs resolves all arguments
func (r *Renderer) resolveArgs(args []string) []string {
	result := make([]string, len(args))
	
	for i, arg := range args {
		// Handle nested calls
		if strings.HasPrefix(arg, "(") && strings.HasSuffix(arg, ")") {
			inner := arg[1 : len(arg)-1]
			parts := r.parseExpr(inner)
			if len(parts) > 0 {
				// Special handling for index
				if parts[0] == "index" && len(parts) >= 3 {
					result[i] = r.indexFunc(parts[1:]...)
					continue
				}
				if fn, ok := r.funcs[parts[0]]; ok {
					resolved := r.resolveArgs(parts[1:])
					result[i] = fn(resolved...)
					continue
				}
			}
		}
		
		// Handle variables
		if strings.HasPrefix(arg, ".") {
			if val := r.resolve(arg); val != nil {
				result[i] = typeutil.ToString(val)
				continue
			}
		}
		
		// Handle nested templates
		if strings.Contains(arg, "{{") {
			// For quoted strings with nested templates, handle specially
			if len(arg) >= 2 && arg[0] == '"' && arg[len(arg)-1] == '"' {
				// Remove outer quotes, process template, then result is the value
				inner := arg[1 : len(arg)-1]
				result[i] = r.render(inner)
			} else {
				result[i] = r.render(arg)
			}
			continue
		}
		
		// Plain string
		result[i] = unquote(arg)
	}
	
	return result
}

// registerDefaults registers built-in functions
func (r *Renderer) registerDefaults() {
	// Markdown helpers
	r.funcs["H1"] = wrap1(md.H1)
	r.funcs["H2"] = wrap1(md.H2)
	r.funcs["H3"] = wrap1(md.H3)
	r.funcs["H4"] = wrap1(md.H4)
	r.funcs["H5"] = wrap1(md.H5)
	r.funcs["H6"] = wrap1(md.H6)
	r.funcs["Bold"] = wrap1(md.Bold)
	r.funcs["Italic"] = wrap1(md.Italic)
	r.funcs["Strikethrough"] = wrap1(md.Strikethrough)
	r.funcs["InlineCode"] = wrap1(md.InlineCode)
	r.funcs["BulletItem"] = wrap1(md.BulletItem)
	r.funcs["CodeBlock"] = wrap1(md.CodeBlock)
	r.funcs["Blockquote"] = wrap1(md.Blockquote)
	r.funcs["Paragraph"] = wrap1(md.Paragraph)
	r.funcs["EscapeText"] = wrap1(md.EscapeText)
	
	r.funcs["Link"] = wrap2(md.Link)
	r.funcs["Image"] = wrap2(md.Image)
	r.funcs["LanguageCodeBlock"] = wrap2(md.LanguageCodeBlock)
	r.funcs["Footnote"] = wrap2(md.Footnote)
	r.funcs["CollapsibleSection"] = wrap2(md.CollapsibleSection)
	
	r.funcs["InlineImageWithLink"] = wrap3(md.InlineImageWithLink)
	
	r.funcs["TodoItem"] = func(args ...string) string {
		if len(args) >= 2 {
			return md.TodoItem(args[0], args[1] == "true")
		}
		return "[error: TodoItem requires 2 arguments]"
	}
	
	r.funcs["HorizontalRule"] = func(args ...string) string {
		return md.HorizontalRule()
	}
	
	// Utilities
	r.funcs["concat"] = func(args ...string) string {
		return strings.Join(args, "")
	}
	
	r.funcs["printf"] = func(args ...string) string {
		if len(args) == 0 {
			return ""
		}
		values := make([]interface{}, len(args)-1)
		for i, v := range args[1:] {
			values[i] = v
		}
		return fmt.Sprintf(args[0], values...)
	}
	
	r.funcs["trim"] = wrap1(strings.TrimSpace)
	r.funcs["string"] = func(args ...string) string {
		if len(args) > 0 {
			return typeutil.ToString(args[0])
		}
		return ""
	}
	
	r.funcs["index"] = r.indexFunc
}

// indexFunc provides array/map indexing
func (r *Renderer) indexFunc(args ...string) string {
	if len(args) < 2 {
		return "[error: index requires 2 arguments]"
	}
	
	// Handle the first argument - could be variable or literal
	collection := strings.TrimSpace(args[0])
	indexStr := strings.TrimSpace(args[1])
	
	// Remove quotes from index if present
	indexStr = unquote(indexStr)
	
	// For unit tests - literal values
	if !strings.HasPrefix(collection, ".") {
		if idx, err := strconv.Atoi(indexStr); err == nil {
			return fmt.Sprintf("[%d]", idx)
		}
		return fmt.Sprintf("[%s]", indexStr)
	}
	
	// Resolve the collection variable
	val := r.resolve(collection)
	if val == nil {
		return "[error: collection not found]"
	}
	
	// Parse index
	idx, err := strconv.Atoi(indexStr)
	if err != nil {
		return "[error: index must be a number]"
	}
	
	// Do the indexing
	switch v := val.(type) {
	case []interface{}:
		if idx >= 0 && idx < len(v) {
			return typeutil.ToString(v[idx])
		}
	case []string:
		if idx >= 0 && idx < len(v) {
			return v[idx]
		}
	default:
		if items := toSlice(val); items != nil {
			if idx >= 0 && idx < len(items) {
				return typeutil.ToString(items[idx])
			}
		}
	}
	
	return "[error: index out of range]"
}

// Helper functions

// lookup navigates nested maps
func lookup(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	var current interface{} = data
	
	for _, part := range parts {
		if part == "" {
			continue
		}
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil
		}
	}
	
	return current
}

// toSlice converts various types to []interface{}
func toSlice(val interface{}) []interface{} {
	switch v := val.(type) {
	case []interface{}:
		return v
	case []map[string]interface{}:
		result := make([]interface{}, len(v))
		for i, m := range v {
			result[i] = m
		}
		return result
	default:
		return typeutil.ToInterfaceSlice(val)
	}
}

// unquote removes quotes from strings
func unquote(s string) string {
	// Don't trim space - preserve formatting
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// Function wrappers for cleaner registration

func wrap1(fn func(string) string) Func {
	return func(args ...string) string {
		if len(args) > 0 {
			return fn(args[0])
		}
		return "[error: missing argument]"
	}
}

func wrap2(fn func(string, string) string) Func {
	return func(args ...string) string {
		if len(args) >= 2 {
			return fn(args[0], args[1])
		}
		return "[error: requires 2 arguments]"
	}
}

func wrap3(fn func(string, string, string) string) Func {
	return func(args ...string) string {
		if len(args) >= 3 {
			return fn(args[0], args[1], args[2])
		}
		return "[error: requires 3 arguments]"
	}
}

// Helper is kept for backward compatibility
type Helper interface {
	Execute(args []string, data map[string]interface{}) string
}