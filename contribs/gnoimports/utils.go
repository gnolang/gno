package main

import "go/ast"

// hasIota checks if a spec contains the built-in "iota" identifier.
func hasIota(s ast.Spec) bool {
	has := false
	ast.Inspect(s, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && id.Name == "iota" && id.Obj == nil {
			has = true
			return false
		}
		return true
	})
	return has
}

// isPredeclared reports whether an identifier is predeclared.
func isPredeclared(s string) bool {
	return predeclaredTypes[s] || predeclaredFuncs[s] || predeclaredConstants[s]
}

var (
	predeclaredTypes = map[string]bool{
		"any":        true,
		"bool":       true,
		"byte":       true,
		"comparable": true,
		"complex64":  true,
		"complex128": true,
		"error":      true,
		"float32":    true,
		"float64":    true,
		"int":        true,
		"int8":       true,
		"int16":      true,
		"int32":      true,
		"int64":      true,
		"rune":       true,
		"string":     true,
		"uint":       true,
		"uint8":      true,
		"uint16":     true,
		"uint32":     true,
		"uint64":     true,
		"uintptr":    true,
	}
	predeclaredFuncs = map[string]bool{
		"append":  true,
		"cap":     true,
		"close":   true,
		"complex": true,
		"copy":    true,
		"delete":  true,
		"imag":    true,
		"len":     true,
		"make":    true,
		"new":     true,
		"panic":   true,
		"print":   true,
		"println": true,
		"real":    true,
		"recover": true,
	}
	predeclaredConstants = map[string]bool{
		"false": true,
		"iota":  true,
		"nil":   true,
		"true":  true,
	}
)
