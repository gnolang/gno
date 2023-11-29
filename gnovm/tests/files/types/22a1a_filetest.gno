package main

type word uint
type nat []word

func (n nat) add() bool {
	return true
}

func main() {
	var abs nat
	abs = []word{0} // unname-Composit to named DeclaredTypes
	println(abs.add())

	var w []word
	w = abs
	println(w)
	switch w.(type) {
	case []word:
		println("[]word")
	case nat:
		println("should not happen")
	}
}

// Output:
// true
// slice[(0 main.word)]
// []word
