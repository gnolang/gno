package main

import "fmt"

func main() {
	x := 2
	s := []int{1 + 1<<x}
	fmt.Println(s)
	fmt.Printf("%T\n", s[0])

	//s1 := []int{0}
	//s1[0] = 1 + 1<<2
	//fmt.Printf("%T\n", s1[0])
}

// Output:
// [5]
// int
