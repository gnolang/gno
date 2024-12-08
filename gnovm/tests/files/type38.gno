package main

import "fmt"

type IntArray []int
type Arr *IntArray

func add(arr Arr) { // receiver is val, not ptr
	*arr = append(*arr, 1)
}

func main() {
	a := new(IntArray)
	add(a)

	fmt.Println(a)
	fmt.Println(*a)
	fmt.Println(len(*a))
	fmt.Println((*a)[0])
}

// Output:
// &[1]
// [1]
// 1
// 1
