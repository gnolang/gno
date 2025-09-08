package main

import "fmt"

func main() {
	mp := map[any]int{nil: 1, 1: 1}
	fmt.Printf("%v", mp)
}

// Output:
// map[<nil>:1 1:1]
