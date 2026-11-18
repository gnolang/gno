package main

// As rune14, in arithmetic (gc: -96).

var s uint

func main() {
	println(1 + ((-'a') << s))
}

// Error:
// main/rune17.gno:8:16-20: unexpected type <untyped> int32 in operation
