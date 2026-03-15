package main

func main() {
	b := make([]byte, "h")

	println(b)
}

// Error:
// main/varg_2.gno:4:7-24: cannot use untyped string as IntKind

// TypeCheckError:
// main/varg_2.gno:4:20: cannot convert "h" (untyped string constant) to type int
