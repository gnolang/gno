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

// Output:
// [4]
