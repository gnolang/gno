package main

import (
	"bytes"
	"std"
)

func main() {
	caller := std.GetOrigCaller()
	caller2 := std.GetOrigCaller()
	cmp := bytes.Compare([]byte(caller), []byte(caller2))
	println(cmp)
}

// Output:
// 0
