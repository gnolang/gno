package main

var a int = 0
var b bool = false

func main() {
	if a == b {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Error:
// main/gno.land/p/demo/challenge/comparison_5_filetest.gno:7: cannot use int as bool
