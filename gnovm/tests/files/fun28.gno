package main

func f(i int) {}

func main() {
	f(nil)
}

// Error:
// main/files/fun28.gno:6:2: cannot use nil as int value in argument to f<VPBlock(3,0)>

// TypeCheckError:
// main/files/fun28.gno:6:4: cannot use nil as int value in argument to f
