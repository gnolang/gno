package main

import "github.com/gnolang/gno/_test/c1"

func main() {
	println(c1.C1)
}

// Error:
// import cycle detected: "github.com/gnolang/gno/_test/c1"
