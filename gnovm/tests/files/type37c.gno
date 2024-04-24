package main

import "fmt"

type Integer int

func (i *Integer) Add(x int) { // receiver is val, not ptr
	println(int(*i) + x)
}

func main() {
	a := new(Integer)
	a.Add(4)

	fmt.Println(*a)
}

// Output:
// 4
// 0
