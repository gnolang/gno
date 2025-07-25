package main

func f(i int) {}

func main() {
	f(nil)
}

// Error:
// main/fun28.gno:6:2-8: cannot use nil as int value in argument to f<VPBlock(3,0)>

// TypeCheckError:
// main/fun28.gno:6:4: cannot use nil as int value in argument to f
