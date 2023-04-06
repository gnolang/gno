package main

import (
	"fmt"
)


type word uint
type nat []word

type Int struct {
	neg bool
	abs nat
}

func (n nat) add() bool {
	return true
}

func main() {
	z := &Int{
		neg: true,
	}

	z.abs = []word{0}
	println(z.abs.add())
}

// Output:
// true