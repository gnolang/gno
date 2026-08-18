package main

func main() {
	var s uint = 33
	var u2 = string(1<<s + 1.0)
	println(u2)
}

// Error:
// main/shift_f4a.gno:5:11-29: cannot use untyped Bigint as StringKind

// TypeCheckError:
// main/shift_f4a.gno:5:18: cannot convert 1 << s + 1.0 (untyped float value) to type string
