// MAXALLOC: 25000
// max total allocation of 100 mb.

package main

import "runtime"

func main() {
	w := "world"
	s := "hello" + w
	runtime.GC()
}
