// Defining from a nil interface conversion is legal: T(nil) carries the
// interface type via ATTR_TYPEOF_VALUE even though its TypedValue is
// undefined — only bare `nil` has no type to infer from.
package main

type zilch any

func main() {
	x := zilch(nil)
	var y = zilch(nil)
	println(x == nil, y == nil)
}

// Output:
// true true
