package main

import "fmt"

func f() int {
	return 2
}

func main() {
	var a, b, c = 1, f(), 3
	fmt.Println(a, b, c)

	x, y, z := 1, f(), 3
	fmt.Println(x, y, z)
}

// Output:
// 1 2 3
// 1 2 3
