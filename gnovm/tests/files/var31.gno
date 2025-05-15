package main

func main() {
	var t = struct{}
}

// Error:
// main/files/var31.gno:4:6: struct{} (type) is not an expression

// TypeCheckError:
// main/files/var31.gno:4:10: struct{} (type) is not an expression; main/files/var31.gno:4:6: declared and not used: t
