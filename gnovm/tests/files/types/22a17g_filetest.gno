package main

import "fmt"

type word uint
type nat []word

// receiver
func (n nat) add() bool {
	return true
}

func Gen() interface{} {
	n := []word{0}
	return n
}

// mapLit
func main() {
	r := Gen()

	fmt.Printf("%T \n", r)

	switch r.(type) {
	case nat:
		println("nat")
	case []word:
		println("[]word")
	default:
		panic("should not happen")
	}
}

// Output:
// []uint
// []word
