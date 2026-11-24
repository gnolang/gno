package main

// The type-switch var is a binding per clause, refused at the deploy gate.

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
