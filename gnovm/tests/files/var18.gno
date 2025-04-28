package main

func main() {
	var a, b, c = 1, 2
}

// Error:
// main/files/var18.gno:4:6: missing init expr for c<!VPUverse(0)>

// TypeCheckError:
// main/files/var18.gno:4:12: missing init expr for c; main/files/var18.gno:4:6: declared and not used: a; main/files/var18.gno:4:9: declared and not used: b; main/files/var18.gno:4:12: declared and not used: c
