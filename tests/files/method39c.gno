package main

import (
	"fmt"
)

type word uint
type nat []word

type Int struct {
	nat
	neg bool
}


func (n nat) add() bool {
	return true
}

func main() {
	z := &Int{
		nat: []word{0},
		neg: true,
	}
	z.add()
	z.nat.add()
}
