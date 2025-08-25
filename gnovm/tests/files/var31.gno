package main

func main() {
	var t = struct{}
}

// Error:
// main/var31.gno:4:6-18: struct{} (type) is not an expression

// TypeCheckError:
// main/var31.gno:4:10: struct{} (type) is not an expression; main/var31.gno:4:6: declared and not used: t
