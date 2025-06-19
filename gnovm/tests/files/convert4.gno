package main

func main() {
	println(int(nil))
}

// Error:
// main/convert4.gno:4:10-18: cannot convert (const (undefined)) to IntKind

// TypeCheckError:
// main/convert4.gno:4:14: cannot convert nil to type int
