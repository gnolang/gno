package main

func main() {
	var t = nil
}

// Error:
// main/files/var32.gno:4:6: use of untyped nil in variable declaration

// TypeCheckError:
// main/files/var32.gno:4:10: use of untyped nil in variable declaration; main/files/var32.gno:4:6: declared and not used: t
