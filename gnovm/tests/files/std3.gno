package main

import (
	"bytes"
	"std"
)

func main() {
	caller := std.OriginCaller()
	caller2 := std.OriginCaller()
	cmp := bytes.Compare([]byte(caller), []byte(caller2))
	println(cmp)
}

// Output:
// 0
