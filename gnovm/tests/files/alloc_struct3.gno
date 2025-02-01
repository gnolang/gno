// MAXALLOC: 50000
package main

import "runtime"

type Bar struct {
	text string
}

type Foo struct {
	name string
	b    Bar
	//s    []Bar
}

//var S = []Bar{{"a"}, {"b"}}
//
//var bar = Bar{name: "bar"}
//var f3 = Foo{name: "foo3", b: bar, s: S}
//

//var f4 = Foo{name: "foo4", b: Bar{name: "bar"}, s: S}

//var f5 = Foo{name: "foo4", b: Bar{name: "bar"}, s: []Bar{{"c"}, {"d"}}}

//var f1 = Foo{b: Bar{text: "bar"}}

// alloc firstly while initialization
// alloc secondely while value copy
//var f2 = Foo{name: "f2"}

var s1 = []int{1, 2, 3} // allocate
// var s2 = s1 // no allocate
var s3 = s1[1:]

//var f3 = &Foo{name: "ref"}
//var f33 = f3

func main() {
	runtime.GC()
}
