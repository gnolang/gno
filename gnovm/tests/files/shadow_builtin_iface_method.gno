package main

// Interface method parameter names are checked like a function type's.

type I interface{ M(len int) }

func main() {
	var i I
	_ = i
}

// Error:
// main/shadow_builtin_iface_method.gno:5:20-29: builtin identifiers cannot be shadowed: len
