package main

import "github.com/gnolang/gno/_test/c1"

func main() {
	println(c1.C1)
}

// Error:
// github.com/gnolang/gno/_test/c2/c2.gno:1: import cycle detected: "github.com/gnolang/gno/_test/c1" (through [github.com/gnolang/gno/_test/c1 github.com/gnolang/gno/_test/c2])
