package main

import "fmt"

func main() {
	a := [3]int{2: 2, 1}
	fmt.Println(a)
}

// Error:
// main/array3.gno:6:7-22: index 3 out of bounds [0:3]

// TypeCheckError:
// main/array3.gno:6:20: index 3 is out of bounds (>= 3)
