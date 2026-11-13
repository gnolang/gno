package main

// Unary operators on untyped rune constants are rejected, as before:
// here in a comparison with a non-constant shift (gc: false).

var s uint

func main() {
	println((-'a' << s) < -100)
}

// Error:
// main/rune14.gno:9:11-15: unexpected type <untyped> int32 in operation
