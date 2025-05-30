package main

import "fmt"

type Foo struct {
	age  int
	name string
}

func main() {
	f := Foo{age: 1, name: "a"}
	fmt.Printf("f: %v\n", f)

	f2 := struct {
		age  int
		name string
	}(f)

	fmt.Printf("f2: %v\n", f2)

	f11 := Foo(struct {
		age  int
		name string
	}{age: 1, name: "b"})

	fmt.Printf("f11: %v\n", f11)
}

// Output:
// f: {1 a}
// f2: {1 a}
// f11: {1 b}
