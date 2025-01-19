package main

import "runtime"

type Foo struct {
	name string
}

var f = Foo{name: "foo"} // allocate2
var pf = &Foo{name: "foo2"}

func main() {
	f1 := f  // allocate2
	f2 := pf // no allocate2
	runtime.GC()
}
