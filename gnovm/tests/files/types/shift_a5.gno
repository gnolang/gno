package main

// TODO: support this?
func main() {
	println(1.0 << 1)
	println(1.0 >> 1)
}

// Error:
// main/files/types/shift_a5.gno:5:10: operator << not defined on: BigdecKind
