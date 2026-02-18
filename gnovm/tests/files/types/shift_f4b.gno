package main

func main() {
	var s uint = 33
	var u2 = uint64(1<<s + 1.2)
	println(u2)
}

// Error:
// main/shift_f4b.gno:5:11-29: cannot convert untyped bigdec to integer -- 1.2 not an exact integer

// TypeCheckError:
// main/shift_f4b.gno:5:25: 1.2 (untyped float constant) truncated to uint64
