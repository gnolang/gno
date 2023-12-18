package main

import "fmt"

type word uint
type nat []word

func (n nat) add() bool {
	return true
}

// TODO: seems fmt prints the underlying type, this is another problem to fix.
func main() {
	var abs nat
	abs = []word{0} // unname-Composit to named DeclaredTypes
	fmt.Printf("%T \n", abs)

	var s []word
	s = abs
	fmt.Printf("%T \n", s)
}

// Output:
// []uint
// []uint
