package main

func main() {
	f := func(len int) int { return len }
	println(f(1))
}

// Error:
// main/shadow_builtin_closure.gno:4:7-39: builtin identifiers cannot be shadowed: len
