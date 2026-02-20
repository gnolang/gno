package main

func main() {
	b := make([]byte, "h")

	println(b)
}

// Error:
// cannot convert StringKind to IntKind

// TypeCheckError:
// main/shift_k2.gno:4:20: cannot convert "h" (untyped string constant) to type int
