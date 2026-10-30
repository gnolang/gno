package main

func f() (len int) { return 1 }

func main() {
	println(f())
}

// Output:
// 1

// TypeCheckError:
// main/shadow_builtin_result.gno:3:11: builtin identifiers cannot be shadowed: len
