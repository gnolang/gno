package main

import "fmt"

type Foo struct {
}

func (f Foo) Println(s ...string) {
	println(s...)
}

func main() {
	println("hello")
	f := Foo{}
	defer func() {
		println("bye1")
	}()
	defer println()
	defer println("p1")
	defer println("p2", "p3")
	defer f.Println()
	defer f.Println("f1")
	defer f.Println("f2", "f3")
	defer fmt.Println()
	defer fmt.Println("fmt1")
	defer fmt.Println("fmt2", "fmt3")
	println("world")
}

// Output:
// hello
// world
// fmt2 fmt3
// fmt1
//
// f2 f3
// f1
//
// p2 p3
// p1
//
// bye1

// TypeCheckError:
// main/defer0.gno:9:11: invalid operation: invalid use of ... with built-in println
