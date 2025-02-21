// MAXALLOC: 50000
package main

import "runtime"

type Foo struct{ name string }

type Bar struct{ num int }

func (f *Foo) Foo() {
	defer func(args ...Bar) { // defer varg
		println(args)
		//runtime.GC() // XXX,GC here does not reclaim memory, but to observe vargs
	}(Bar{1}, Bar{2}, Bar{3})

	println(f.name)
	runtime.GC()
}

var f = Foo{name: "foo"}

func main() {
	f.Foo()
}

// Output:
