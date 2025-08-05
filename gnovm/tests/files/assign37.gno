package main

func f() { }

func main() {
	a := f()
}

// Error:
// main/assign37.gno:6:2-10: f<VPBlock(3,0)> (no value) used as value

// TypeCheckError:
// main/assign37.gno:6:7: f() (no value) used as value; main/assign37.gno:6:2: declared and not used: a
