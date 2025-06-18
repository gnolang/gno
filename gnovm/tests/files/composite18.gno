package main

import "fmt"

func main() {
	x := 0
	a := map[int]int{x: 10, -1 * x: 20}
	fmt.Println(a)
}

// Output:
// map[0:20]
