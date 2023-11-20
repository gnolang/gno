package main

var a int = 0
var b int64 = 0

func main() {
	// suppose compile fail
	if a == b {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Error:
// main/gno.land/p/demo/challenge/comparison_4_filetest.gno:8: mismatch type for EQL or NEQ
