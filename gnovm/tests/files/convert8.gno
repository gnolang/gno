package main

// This filetest tests that dereferencing after converting to a pointer type of
// a re-declared type results in the conversion elem type.

import "fmt"

type Foo struct{ A int }

func (f Foo) String() string { return "foo" }

type Bar Foo

func (b Bar) String() string { return "bar" }

func main() {
	a := &Foo{A: 1}
	b := (*Bar)(a)
	c := any(nil)
	c = *b
	fmt.Println("should be bar:", c)
}

// Output:
// should be bar: bar
