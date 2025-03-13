package main

import "fmt"

func check(v1, v2 any) bool {
	return v1 == v2
}

func main() {
	type t1 int
	type t2 int
	v1 := t1(1)
	v2 := t2(1)
	v3 := t2(3)

	fmt.Println("v1, v2", v1, v2, check(v1, v2))
	fmt.Println("v1, v3", v1, v3, check(v1, v3))
	fmt.Println("v2, v3", v2, v3, check(v2, v3))
}

// Output:
// v1, v2 1 1 false
// v1, v3 1 3 false
// v2, v3 1 3 false
