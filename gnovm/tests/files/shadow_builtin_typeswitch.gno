package main

func main() {
	switch len := any(1).(type) {
	case int:
		println(len)
	}
}

// Error:
// main/shadow_builtin_typeswitch.gno:5:2-6:15: builtin identifiers cannot be shadowed: len
