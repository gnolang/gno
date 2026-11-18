package main

// As rune14, in an explicit conversion (gc: -97).

var s uint

func main() {
	println(int64(-'a' << s))
}

// Error:
// main/rune18.gno:8:16-20: unexpected type <untyped> int32 in operation
