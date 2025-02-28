// MAXALLOC: 100000000
package main

import "runtime"

type Err struct {
	s string
}

func main() {
	defer func() {
		runtime.GC()
	}()

	panic(Err{"user panic"})
}

// Error:
// (struct{("user panic" string)} main.Err)
