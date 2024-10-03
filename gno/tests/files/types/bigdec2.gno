package main

var a uint64

func main() {
	a = 1
	println(a % 1.2)
}

// Error:
// main/files/types/bigdec2.gno:7:10: cannot convert untyped bigdec to integer -- 1.2 not an exact integer
