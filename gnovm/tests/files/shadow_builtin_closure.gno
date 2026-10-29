package main

func main() {
	f := func(len int) int { return len }
	println(f(1))
}

// Output:
// 1

// TypeCheckError:
// main/shadow_builtin_closure.gno:4:12: builtin identifiers cannot be shadowed: len
