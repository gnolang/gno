package main

func main() {
	var s uint = 33
	var u2 = 1<<s != 1.0<<s
	println(u2)
}

// Error:
// main/files/types/shift_f2d.gno:5:11: incompatible types in binary expression: <untyped> bigint NEQ <untyped> bigdec
