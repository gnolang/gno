package main

func fn() {}

func main() {
	// &fn refers to the package func; the local declared after it does not
	// shadow this use.
	_ = &fn
	fn := 1
	_ = fn
}

// Error:
// main/addressable_12g_err.gno:8:6-9: invalid operation: cannot take address of fn<VPBlock(3,0)> (value of type func())

// TypeCheckError:
// main/addressable_12g_err.gno:8:7: invalid operation: cannot take address of fn (value of type func())
