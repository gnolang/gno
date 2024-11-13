package main

func main() {
	println(*nil)
}

// Error:
// main/files/ptr10.gno:4:10: invalid operation: cannot indirect nil
