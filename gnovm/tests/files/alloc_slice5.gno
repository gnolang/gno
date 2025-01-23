// MAXALLOC: 100

package main

import "runtime"

var S = []string{"a", "b", "c"}
var S2 = S

var f = func() { println("func lit") } // 168

var SS = "hello" + "world"

type Foo struct{ name string }

func main() {
	runtime.GC()
}

// Output:
//
