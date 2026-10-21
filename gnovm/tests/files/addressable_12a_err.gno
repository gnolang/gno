package main

func bar() {}

func main() {
	_ = &bar
}

// TypeCheckError:
// main/addressable_12a_err.gno:6:7: invalid operation: cannot take address of bar (value of type func())

// Error:
// main/addressable_12a_err.gno:6:6-10: invalid operation: cannot take address of bar<VPBlock(3,0)> (value of type func())
