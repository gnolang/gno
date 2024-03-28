package main

type c uint
type word c
type nat []word

func (n nat) add() bool {
	return true
}

func main() {
	var abs nat

	abs = []uint{0}

	println(abs.add())
}

// Error:
// main/files/declaredType6b.gno:14: cannot use uint as main.word without explicit conversion