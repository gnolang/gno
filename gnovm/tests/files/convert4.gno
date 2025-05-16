package main

func main() {
	println(int(nil))
}

// Error:
// main/files/convert4.gno:4:10: cannot convert (const (undefined)) to IntKind

// TypeCheckError:
// main/files/convert4.gno:4:14: cannot convert nil to type int
