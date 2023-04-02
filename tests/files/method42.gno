package main

import (
	"fmt"
)

type nat3 []Word
type Word uint


func (n nat3) add3(x, y nat3) bool {
	return true
}

func main() {
	var abs nat3
	abs = []Word{0}
	x := []Word{1}
	y := []Word{2}
	abs.add3(x, y)
}

