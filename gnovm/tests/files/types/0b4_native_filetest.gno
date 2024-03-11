package main

import (
	"errors"
)

func main() {
	errCmp := errors.New("xxx")
	println(5 == errCmp)
}

// Error:
// main/files/types/0b4_native_filetest.gno:9: unexpected type pair: cannot use <untyped> bigint as gonative{error}
