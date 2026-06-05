package main

func main() {
	b := make([]byte, "h")

	println(b)
}

// Error:
// main/varg_2.gno:4:7-24: cannot use ("h" <untyped> string) as type int in argument to make

// TypeCheckError:
// main/varg_2.gno:4:20: cannot convert "h" (untyped string constant) to type int
