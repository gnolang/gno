package main

// A function type is not a binding, but its parameter and result names are
// checked too, so the rule has no exception.

var f func(len int) (cap int)

func main() {
	_ = f
}

// TypeCheckError:
// main/shadow_builtin_functype.gno:6:12: builtin identifiers cannot be shadowed: len
