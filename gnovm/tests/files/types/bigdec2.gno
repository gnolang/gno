package main

var a uint64

func main() {
	a = 1
	println(a % 1.2)
}

// Error:
// main/bigdec2.gno:7:10-17: cannot convert untyped bigdec to integer -- 1.2 not an exact integer

// TypeCheckError:
// main/bigdec2.gno:7:14: 1.2 (untyped float constant) truncated to uint64
