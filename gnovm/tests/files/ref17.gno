package main

import "fmt"

// Deref of a converted pointer type with primitive underlying type.

type Meters int
type Feet int

func main() {
	f := Feet(3)
	v := *((*Meters)(&f))
	fmt.Printf("%T %v\n", v, v)
}

// Output:
// main.Meters 3
