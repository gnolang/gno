package main

// This test is different than var19.b.
// Do not let the linter remove `var`.

func main() {
	var a, b, c = 1, a + 1
	println(a)
	println(b)
	println(c)
}

// Error:
// main/var19b.gno:7:6-24: missing init expr for c<!VPInvalid(0)>

// TypeCheckError:
// main/var19b.gno:7:12: missing init expr for c; main/var19b.gno:7:19: undefined: a
