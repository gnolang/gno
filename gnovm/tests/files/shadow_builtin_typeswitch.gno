package main

// The type-switch var is Reserved, and checked, in each clause.

func main() {
	switch len := any(1).(type) {
	case int:
		println(len)
	}
}

// Output:
// 1

// TypeCheckError:
// main/shadow_builtin_typeswitch.gno:6:9: builtin identifiers cannot be shadowed: len
