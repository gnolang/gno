package main

// The type-switch var is Reserved, and checked, in each clause.

func main() {
	switch len := any(1).(type) {
	case int:
		println(len)
	}
}

// Error:
// main/shadow_builtin_typeswitch.gno:7:2-8:15: builtin identifiers cannot be shadowed: len
