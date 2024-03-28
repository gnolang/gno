package main

type word uint
type nat map[string]word

func (n nat) add() bool {
	return true
}

func main() {
	var abs nat
	abs = map[string]word{"test": word(0)} // unname-Composit to named DeclaredTypes

	println(abs.add())
}

// Output:
// true