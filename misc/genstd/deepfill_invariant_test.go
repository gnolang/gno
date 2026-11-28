package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDeepFillOnlyOnPointerFreeParams pins the invariant behind leaving
// Value.DeepFill without a cycle detector.
//
// The generator emits tvN.DeepFill only for parameters that are not
// gno.TypedValue (template.tmpl). Every such parameter today is a scalar or a
// slice of scalars, and none of those can reference the heap item that a
// self-referential pointer (`var x any; x = &x`) creates. So DeepFill never
// meets the cyclic graph the fillValueTV guard is about.
//
// A parameter type that can hold such a reference (a pointer, struct,
// interface, map, func, or slice of one) would put the fatal, uncatchable Go
// stack overflow back: DeepFill recurses through PointerValue.Base and
// HeapItemValue.Value with no visited set. Extend the allowlist only after
// giving DeepFill a cycle detector of its own.
func TestDeepFillOnlyOnPointerFreeParams(t *testing.T) {
	// Scalar and scalar-slice Go types, none of which can reach a heap item.
	pointerFree := map[string]bool{
		"bool":     true,
		"string":   true,
		"rune":     true,
		"byte":     true,
		"int":      true,
		"int8":     true,
		"int16":    true,
		"int32":    true,
		"int64":    true,
		"uint":     true,
		"uint8":    true,
		"uint16":   true,
		"uint32":   true,
		"uint64":   true,
		"float32":  true,
		"float64":  true,
		"[]byte":   true,
		"[]int64":  true,
		"[]string": true,
	}

	for _, dir := range []string{"../../gnovm/stdlibs", "../../gnovm/tests/stdlibs"} {
		t.Run(dir, func(t *testing.T) {
			chdir(t, dir)

			pkgs, err := walkStdlibs(".")
			require.NoError(t, err)

			for _, m := range linkFunctions(pkgs) {
				for i, p := range m.Params {
					if p.IsTypedValue {
						continue // template.tmpl skips DeepFill for these
					}
					if !pointerFree[p.GoQualifiedName()] {
						t.Errorf("%s.%s param %d has type %q; the generator calls DeepFill on it and it may reference a heap item. Guard DeepFill, or add the type to the allowlist only after confirming it cannot",
							m.GnoImportPath, m.GnoFunc, i, p.GoQualifiedName())
					}
				}
			}
		})
	}
}
