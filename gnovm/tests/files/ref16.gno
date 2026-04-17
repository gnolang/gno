package main

import "fmt"

// Deref of a converted pointer type.
// *((*Foo)(&Bar{})) should produce Foo, not Bar.

type Foo struct{ X int }
type Bar struct{ X int }

func main() {
	b := Bar{X: 1}
	v := *((*Foo)(&b))
	fmt.Printf("%T %v\n", v, v)
}

// Output:
// main.Foo {1}
