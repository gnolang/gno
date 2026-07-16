package main

import (
	"bytes"
	"chain/runtime/unsafe"
)

func main() {
	caller := unsafe.OriginCaller()
	caller2 := unsafe.OriginCaller()
	cmp := bytes.Compare([]byte(caller), []byte(caller2))
	println(cmp)
}

// Output:
// 0
