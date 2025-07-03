package main

func f() {}

func main() {
	var a, b int = f(), 1
}

// Error:
// main/var34c.gno:6:6-23: f<VPBlock(3,0)> (no value) used as value

// TypeCheckError:
// main/var34c.gno:6:17: f() (no value) used as value; main/var34c.gno:6:6: declared and not used: a; main/var34c.gno:6:9: declared and not used: b
