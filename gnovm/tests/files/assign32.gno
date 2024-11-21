package main

func f2() (string, int) {
	return "hey", 0
}

func main() {
	n := f2()
}

// Output:
// main/gnovm/tests/files/assign32.gno:7:1: f2() (2 values) used as single value