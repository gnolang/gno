package main

import (
	"bytes"
	"chain/runtime"
)

func main() {
	caller := runtime.OriginCaller()
	caller2 := runtime.OriginCaller()
	cmp := bytes.Compare([]byte(caller), []byte(caller2))
	println(cmp)
}

// Output:
// 0
