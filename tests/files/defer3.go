package main

import (
	"bytes"
)

func main() {
	println("hello")
	buf := bytes.NewBuffer(nil)
	defer buf.Reset()
}

// Output:
// hello
