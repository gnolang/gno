package main

type c uint
type word c
type nat []word

func (n nat) add() bool {
	return true
}

func main() {
	var abs nat

	abs = []c{0}

	println(abs.add())
}

// Error:
// main/files/declaredType6.gno:14: cannot use main.c as main.word with implicit conversion