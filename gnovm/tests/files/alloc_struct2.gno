package main

import "runtime"

type Foo struct {
	name string
	b    Bar
}

type Bar struct {
	name string
}

var bar := Bar{name:"bar"}
var f3 = Foo{name: "foo1", b: bar}

var f1 = Foo{name: "foo1", b: Bar{name: "bar1"}}

var m = map[string]Bar{"a": Bar{name: "bar2"}}
var S = []Bar{Bar{name: "bar3"}, Bar{name: "bar4"}, Bar{name: "bar5"}}

func main() {
	f2 := &Foo{name: "foo2"} // no allocate2

	f11 := &f1

	b := f1.b
	b2 := m["a"]

	s := S[1:] // new slice
	runtime.GC()
}
