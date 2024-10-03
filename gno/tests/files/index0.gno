package main

import (
	"fmt"
)

func main() {
	a := int8(0)
	b := uint16(2)
	x := [4]int{}
	fmt.Println(x[a:b])
}

// Output:
// [0 0]
