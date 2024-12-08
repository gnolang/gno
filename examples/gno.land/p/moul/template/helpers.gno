package template

import (
	"strings"
)

// Helper interface defines the contract for template helpers
type Helper interface {
	Execute(arg string, data map[string]interface{}) string
}

// SingleArgHelper is a helper that takes a single argument
type SingleArgHelper struct {
	fn func(string) string
}

func (h SingleArgHelper) Execute(arg string, data map[string]interface{}) string {
	// Handle variable resolution first
	if strings.HasPrefix(arg, ".") {
		key := strings.TrimPrefix(arg, ".")
		if val, ok := data[key]; ok {
			return h.fn(convertToString(val))
		}
		return h.fn(arg)
	}

	// Handle quoted strings
	if strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"") {
		// Remove outer quotes but preserve escaped quotes
		arg = strings.TrimPrefix(strings.TrimSuffix(arg, "\""), "\"")
		return h.fn(strings.TrimSpace(arg))
	}

	return h.fn(strings.TrimSpace(arg))
}

// TwoArgHelper is a helper that takes two arguments
type TwoArgHelper struct {
	fn func(string, string) string
}

func (h TwoArgHelper) Execute(args string, data map[string]interface{}) string {
	parts := splitParts(args)
	if len(parts) < 2 {
		return "{{ invalid two-arg helper }}"
	}

	arg1 := resolveArg(parts[0], data)
	arg2 := resolveArg(parts[1], data)
	return h.fn(arg1, arg2)
}

// VarArgHelper is a helper that takes variable number of arguments
type VarArgHelper struct {
	fn func([]string) string
}

func (h VarArgHelper) Execute(args string, data map[string]interface{}) string {
	parts := splitParts(args)
	resolved := make([]string, len(parts))
	for i, part := range parts {
		resolved[i] = resolveArg(part, data)
	}
	return h.fn(resolved)
}

// NewConcatHelper for string concatenation with variable arguments
func NewConcatHelper() Helper {
	return VarArgHelper{
		fn: func(args []string) string {
			return strings.Join(args, "")
		},
	}
}

// Helper utility functions
func splitParts(s string) []string {
	var parts []string
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
				parts = append(parts, current.String())
				current.Reset()
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

// resolveArg resolves an argument, handling both quoted strings and variables
func resolveArg(arg string, data map[string]interface{}) string {
	// Handle variables
	if strings.HasPrefix(arg, ".") {
		key := strings.TrimPrefix(arg, ".")
		if val, ok := data[key]; ok {
			return convertToString(val)
		}
		return arg
	}

	// Handle quoted strings
	if strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"") {
		arg = strings.Trim(arg, "\"")
		// Handle escaped quotes properly
		if strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"") {
			arg = strings.Trim(arg, "\"")
		}
		arg = strings.ReplaceAll(arg, "\\\"", "\"")
		return arg
	}

	return arg
}
