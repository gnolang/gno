package main

import "fmt"

type IntArray []int
type Arr *IntArray

func (a Arr) Add(x int) { // receiver is val, not ptr
	*a = append(*a, x)
}

func main() {
	a := new(IntArray)
	Arr(a).Add(4)

	fmt.Println(*a)
}

// Error:
// main/files/type37.gno:8: invalid receiver type main.Arr (base type is pointer type)
