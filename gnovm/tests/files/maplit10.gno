package main

import "fmt"

func main() {
	a := map[[2]int]int{[2]int{1, 2}: 10, [2]int{3, 4}: 20}
	fmt.Println(a)
}

// Output:
// map[[1 2]:10 [3 4]:20]
