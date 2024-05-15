package main

import (
	"io"
)

type reader struct{}

/*
func (r reader) Read(p []byte) (n int, err error) {
	return
}
*/

func main() {
	var r io.Reader
	i := reader{}
	r = i
	println(r)
}

// Output:
// (struct{} main.reader)
