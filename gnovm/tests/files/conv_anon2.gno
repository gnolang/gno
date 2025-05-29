package main

import "fmt"

type Foo struct {
	age  int
	name string
}

var Arr = [2]Foo{Foo{age: 1, name: "a"}, Foo{age: 2, name: "b"}}

func main() {
	Arr2 := [2]struct {
		age  int
		name string
	}(Arr)

	fmt.Println("Arr2: ", Arr2)
}

// Error:
// main/conv_anon2.gno:13:10-16:8: cannot use [2]main.Foo as [2]struct{age int; name string}

// TypeCheckError:
// main/conv_anon2.gno:16:4: cannot convert Arr (variable of type [2]Foo) to type [2]struct{age int; name string}
