// MAXALLOC: 25000
// max total allocation of 100 mb.

package main

import "runtime"

func main() {
	var s = "hello" // const
	s += "world"
	runtime.GC()
}
