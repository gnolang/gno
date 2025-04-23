package main

func f() {}

func main() {
	var t = f()
}

// Error:
// main/files/var34.gno:6:6: f<VPBlock(3,0)> (no value) used as value

// TypeCheckError:
// main/files/var34.gno:6:10: f() (no value) used as value; main/files/var34.gno:6:6: declared and not used: t
