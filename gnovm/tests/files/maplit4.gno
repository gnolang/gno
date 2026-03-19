package main

import "fmt"

func main() {
	i3 := 0
	a := map[int]int{0: 10, 0: 11, i3: 20}
	fmt.Println(a)
}

// Error:
// main/maplit4.gno:7:10-43: duplicate key (0 int) in map literal

// TypeCheckError:
// main/maplit4.gno:7:29: duplicate key 0 in map literal
