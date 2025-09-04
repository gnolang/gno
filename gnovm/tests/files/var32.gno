package main

func main() {
	var t = nil
}

// Error:
// main/var32.gno:4:6-13: use of untyped nil in variable declaration

// TypeCheckError:
// main/var32.gno:4:10: use of untyped nil in variable declaration; main/var32.gno:4:6: declared and not used: t
