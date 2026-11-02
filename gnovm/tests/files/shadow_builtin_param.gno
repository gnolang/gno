package main

// Parameters are bindings; shadowing a builtin there was accepted before.

func f(len int) int { return len }

func main() {
	println(f(1))
}

// Output:
// 1

// TypeCheckError:
// main/shadow_builtin_param.gno:5:8: builtin identifiers cannot be shadowed: len
