package main

func main() {
	println(*nil)
}

// Error:
// main/ptr10.gno:4:10-14: invalid operation: cannot indirect nil

// TypeCheckError:
// main/ptr10.gno:4:11: invalid operation: cannot indirect nil
