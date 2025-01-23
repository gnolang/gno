// MAXALLOC: 25000
// max total allocation of 100 mb.

package main

import "runtime"

func main() {
	s := "hello"
	runtime.GC()
}
