package main

type (
	c    uint
	word c
	nat  []word
)

func (n nat) add() bool {
	return true
}

func main() {
	var abs nat

	abs = []uint{0}

	println(abs.add())
}

// Error:
// main/files/assign_unnamed_type/declaredtype_method/declaredType6b_filetest.gno:16: cannot use uint as main.word without explicit conversion
