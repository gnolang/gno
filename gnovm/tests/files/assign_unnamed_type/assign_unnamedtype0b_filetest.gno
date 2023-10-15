package main

type nat []int
type word int

func main() {
	var a nat
	b := []word{0}
	a = b

	println(a)
}

// Error:
// main/files/assign_unnamed_type/assign_unnamedtype0b_filetest.gno:9: cannot use main.word as int without explicit conversion
