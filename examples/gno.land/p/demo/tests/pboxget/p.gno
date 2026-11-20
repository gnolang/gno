// Package pboxget returns a boxed value from a slice, for view-only copy tests.
package pboxget

// First returns xs[0]; the define copies a struct held in the box.
func First(xs []any) any {
	x := xs[0]
	return x
}
