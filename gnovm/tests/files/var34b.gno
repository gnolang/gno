package main

func f() {}

func main() {
	var a int
	a = f()
}

// Error:
// main/var34b.gno:7:2-9: f<VPBlock(3,0)> (no value) used as value

// TypeCheckError:
// main/var34b.gno:7:6: f() (no value) used as value; main/var34b.gno:6:6: declared and not used: a
