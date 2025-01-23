// MAXALLOC: 25000
// max total allocation of 100 mb.

package main

import "runtime"

var s = "hello"

func main() {
	s += "world" // new string
	s += "!!!"   // new string
	runtime.GC()
}
