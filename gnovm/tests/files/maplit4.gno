package main

import "fmt"

func main() {
	i3 := 0
	a := map[int]int{0: 10, 0: 11, i3: 20}
	fmt.Println(a)
}

// Error:
// main/maplit4.gno:7:7-40: duplicate key (0 int) in map literal

// TypeCheckError:
// main/maplit4.gno:7:26: duplicate key 0 in map literal
