package main

func main() {
	var s uint = 33

	var u1 = 1.0<<s != 0
	println(u1)
}

// Error:
// main/files/types/shift_f1.gno:6:6: operator << not defined on: BigdecKind
