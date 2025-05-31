package main

func main() {
	var a, b, c = 1, 2
}

// Error:
// main/var18.gno:4:6-20: missing init expr for c<!VPInvalid(0)>

// TypeCheckError:
// main/var18.gno:4:12: missing init expr for c; main/var18.gno:4:6: declared and not used: a; main/var18.gno:4:9: declared and not used: b; main/var18.gno:4:12: declared and not used: c
