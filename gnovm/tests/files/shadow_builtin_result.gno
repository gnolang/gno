package main

func f() (len int) { return 1 }

func main() {
	println(f())
}

// Error:
// main/shadow_builtin_result.gno:3:1-32: builtin identifiers cannot be shadowed: len
