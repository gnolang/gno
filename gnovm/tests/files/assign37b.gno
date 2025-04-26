package main

import "fmt"

func f() { }

func main() {
	a, b := f(), f()
}

// Error:
// main/files/assign37b.gno:8:2: f<VPBlock(3,0)> (no value) used as value

// TypeCheckError:
// main/files/assign37b.gno:8:10: f() (no value) used as value; main/files/assign37b.gno:8:15: f() (no value) used as value; main/files/assign37b.gno:8:2: declared and not used: a; main/files/assign37b.gno:8:5: declared and not used: b; main/files/assign37b.gno:3:8: "fmt" imported and not used
