package main

import (
	"fmt"
)

type nat3 []Word
type Word uint


func (n nat3) add3() bool {
	return true
}

func main() {
	var abs nat3
	abs = []Word{0}
	println(abs.add3())
}

// Output:
// true
