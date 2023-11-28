package main

func main() {
	println(int(0) | int64(1))
}

// Error:
// main/gno.land/p/demo/challenge/comparison_0k_filetest.gno:4: cannot use int as int64
