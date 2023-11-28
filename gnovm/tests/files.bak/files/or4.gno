package main

import "fmt"

func foo() (res [4]int) {
	res[0] |= 1
	res[1] |= 2
	res[2] |= 3
	res[3] |= 4
	return
}

func main() {
	a := foo()
	fmt.Println(a)
}

// Output:
// [1 2 3 4]
