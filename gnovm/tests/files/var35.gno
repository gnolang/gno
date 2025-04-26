package main

func main() {
	var i int = nil
}

// Error:
// main/files/var35.gno:4:6: cannot use nil as int value in variable declaration

// TypeCheckError:
// main/files/var35.gno:4:14: cannot use nil as int value in variable declaration; main/files/var35.gno:4:6: declared and not used: i
