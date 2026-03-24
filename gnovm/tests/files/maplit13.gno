package main

import "fmt"

type Inner struct {
	X int
	Y int
}

type Outer struct {
	A Inner
	B string
}

func main() {
	a := map[Outer]int{
		Outer{Inner{1, 2}, "a"}: 10,
		Outer{Inner{3, 4}, "b"}: 20,
	}
	fmt.Println(a)
}

// Output:
// map[{{1 2} a}:10 {{3 4} b}:20]
