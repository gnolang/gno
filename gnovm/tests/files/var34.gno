package main

func f() {}

func main() {
	var t = f()
}

// Error:
// main/var34.gno:6:6-13: f<VPBlock(3,0)> (no value) used as value

// TypeCheckError:
// main/var34.gno:6:10: f() (no value) used as value; main/var34.gno:6:6: declared and not used: t
