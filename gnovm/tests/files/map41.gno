package main

import "fmt"

func main() {
	m := map[any]int{nil: 0, nil: 1}
	fmt.Printf("%v", m)
}

// Output:
// map[<nil>:1]
