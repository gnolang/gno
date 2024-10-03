package main

import "fmt"

type IntArray []int

func (a *IntArray) Add(x int) { // receiver is val, not ptr
	*a = append(*a, x)
}

func main() {
	a := new(IntArray)
	a.Add(4)

	fmt.Println(*a)
}

// Output:
// [4]
