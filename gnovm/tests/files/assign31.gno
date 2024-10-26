package main

import "fmt"

func f() (int) {
	return 2
}

func main() {
	a, b, c := 1, f(), 3
	fmt.Println(a, b, c)
}

// Output:
// 1 2 3
