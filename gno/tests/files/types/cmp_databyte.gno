package main

import "bytes"

func main() {
	cmp := bytes.Compare([]byte("hello"), []byte("hey"))
	println(cmp)

}

// Output:
// -1
