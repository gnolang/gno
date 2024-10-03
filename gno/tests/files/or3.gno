package main

import "fmt"

func main() {
	a := []int{0, 0, 0, 0}
	a[0] |= 1
	a[1] |= 2
	a[2] |= 3
	a[3] |= 4
	fmt.Println(a)
}

// Output:
// [1 2 3 4]
