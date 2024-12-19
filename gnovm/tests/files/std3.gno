package main

import (
	"bytes"
	"std"
)

func main() {
	caller := std.OrigCaller()
	caller2 := std.OrigCaller()
	cmp := bytes.Compare([]byte(caller), []byte(caller2))
	println(cmp)
}

// Output:
// 0
