package main

import "fmt"

func f() { }

func main() {
	a, b := f(), f()
}

// Error:
// main/assign37b.gno:8:2-18: f<VPBlock(3,0)> (no value) used as value

// TypeCheckError:
// main/assign37b.gno:8:10: f() (no value) used as value; main/assign37b.gno:8:15: f() (no value) used as value; main/assign37b.gno:8:2: declared and not used: a; main/assign37b.gno:8:5: declared and not used: b; main/assign37b.gno:3:8: "fmt" imported and not used
