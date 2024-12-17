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
// main/files/assign_unnamed_type/method/declaredType6b_filetest.gno:15:2: cannot use []uint as []main.word
