package main

var a uint64

func main() {
	a = 1.2
	println(a)
}

// Error:
// main/debug/bigdec5.gno:6: cannot convert untyped bigdec to integer -- 1.2 not an exact integer
