package main

type (
	nat  []int
	word int
)

func main() {
	var a nat
	b := []word{0}
	a = b

	println(a)
}

// Error:
// main/files/assign_unnamed_type/unnamedtype0b_filetest.gno:11:2: cannot use main.word as int without explicit conversion
