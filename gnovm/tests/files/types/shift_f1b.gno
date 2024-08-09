package main

func main() {
	var s uint = 33

	var u1 bool
	u1 = 1.0<<s != 0
	println(u1)
}

// Error:
// main/files/types/shift_f1b.gno:7:2: operator << not defined on: BigdecKind
