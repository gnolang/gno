package main

import "fmt"

type Foo struct {
	age  int
	name string
}

var Arr = [2]interface{}{struct {
	age  int
	name string
}{age: 1, name: "a"}, struct {
	age  int
	name string
}{age: 2, name: "b"}}

func main() {
	Arr2 := [2]Foo(Arr)

	fmt.Println("Arr2: ", Arr2)
}

// Error:
// main/convert9a.gno:19:10-21: cannot use [2]interface {} as [2]main.Foo

// TypeCheckError:
// main/convert9a.gno:19:17: cannot convert Arr (variable of type [2]interface{}) to type [2]Foo
