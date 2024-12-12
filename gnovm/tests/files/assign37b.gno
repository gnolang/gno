package main

import "fmt"

func f() { }

func main() {
	a, b := f(), f()
}

// Error:
// main/files/assign37b.gno:8:2: f<VPBlock(3,0)> (no value) used as value
