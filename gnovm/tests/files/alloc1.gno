// MAXALLOC: 25000
// max total allocation of 100 mb.
package main

var c = 1
var b = "hello" // rhs is const
var m = map[string]int{"a": 1}

type Foo struct{ name string }

func bar() {
	f := &Foo{"foo"}
	println(f)
	bb := &Foo{"bar"}
}

func main() {
	a := 1
	bar()
}
