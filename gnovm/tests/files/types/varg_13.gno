package main

func main() {
	_ = make([]int, "hello")
}

// TypeCheckError:
// main/varg_13.gno:4:18: cannot convert "hello" (untyped string constant) to type int

// Error:
// main/varg_13.gno:4:6-26: cannot use ("hello" <untyped> string) as type int in argument to make
