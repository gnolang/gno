package main

import (
	"fmt"
)


type nat3 []Word
type Word uint

type Int3 struct {
	neg bool
	abs nat3
}

func (n nat3) add3() bool {
	return true
}

func main() {
	z := &Int3{
		neg: true,
		abs: []Word{0},
	}
	z.abs.add3()
}
