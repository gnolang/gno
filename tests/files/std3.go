package main

import (
	"bytes"
	"std"
)

func main() {
	caller := std.GetCaller()
	caller2 := std.GetCaller()
	cmp := bytes.Compare([]byte(caller), []byte(caller2))
	println(cmp)
}

// Output:
// 0
