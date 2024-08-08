package main

func main() {
	var s uint = 33
	var u2 = 1.0<<s != 1
	println(u2)
}

// Error:
// main/files/types/shift_f2_0.gno:5:6: operator << not defined on: BigdecKind
