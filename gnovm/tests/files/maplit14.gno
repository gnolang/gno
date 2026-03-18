package main

import "fmt"

func main() {
	a := map[[2][2]int]int{[2][2]int{{1, 2}, {3, 4}}: 10}
	fmt.Println(a)
}

// Output:
// map[[[1 2] [3 4]]:10]
