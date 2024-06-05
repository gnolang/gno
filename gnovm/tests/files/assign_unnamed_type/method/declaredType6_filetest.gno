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
	abs = []c{0}
	println(abs.add())
}

// Error:
// main/files/assign_unnamed_type/method/declaredType6_filetest.gno:15: cannot use main.c as main.word without explicit conversion
