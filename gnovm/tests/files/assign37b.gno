package main

func f() {}

func main() {
	a, b := f(), f()
}

// Error:
// main/assign37b.gno:6:2-18: f<VPBlock(3,0)> (no value) used as value

// TypeCheckError:
// main/assign37b.gno:6:10: f() (no value) used as value; main/assign37b.gno:6:15: f() (no value) used as value; main/assign37b.gno:6:2: declared and not used: a; main/assign37b.gno:6:5: declared and not used: b
