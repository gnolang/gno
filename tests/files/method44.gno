package main

import (
	"fmt"
)

type word uint
type nat []word

func (n nat) add() bool {
	return true
}

// parameter
func main() {
	des := make([]nat, 2)
	src := make([][]word, 2)

	src[0] = []word{0}
	src[1] = []word{1}

	copy(des, src)

	r := des[0]
	println(r.add())

	r = des[1]
	println(r.add())
}

// Output:
// true
// true
