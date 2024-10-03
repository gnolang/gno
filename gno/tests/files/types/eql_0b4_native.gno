package main

import (
	"errors"
)

func main() {
	errCmp := errors.New("xxx")
	println(5 == errCmp)
}

// Error:
// main/files/types/eql_0b4_native.gno:9:10: unexpected type pair: cannot use <untyped> bigint as gonative{error}
