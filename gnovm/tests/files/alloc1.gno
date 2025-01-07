// MAXALLOC: 25000
// max total allocation of 100 mb.
package main

type Foo struct{ name string }

func bar() {
	f := &Foo{"foo"}
	println(f)
}

func main() {
	a := 1
	bar()
}
