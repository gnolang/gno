package main

import "fmt"

type IntArray []int
type Arr *IntArray

func print(arr Arr) { // receiver is val, not ptr
	println(arr)
}

func main() {
	a := new(IntArray)
	print(a)

	fmt.Println(*a)
}

// Output:
// &(nil main.IntArray)
// []
