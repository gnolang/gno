package main

func main() {
	var s uint = 33
	var u2 = 1 + 1.0<<s
	println(u2)
}

// Error:
// main/files/types/shift_f3c.gno:5:6: operator << not defined on: BigdecKind
