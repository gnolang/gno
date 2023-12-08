package main

import "fmt"

type nat []int

func main() {

	var a nat
	a = []int{0}
	b := []int{1}

	fmt.Printf("%v \n ", a)
	fmt.Printf("%v \n ", b)

	a = nat{0}
	b = a

	fmt.Printf("%v \n ", a)
	fmt.Printf("%v \n ", b)

}

// Output:
// [0]
//  [1]
//  [0]
//  [0]
