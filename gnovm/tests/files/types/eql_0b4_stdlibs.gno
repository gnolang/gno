package main

import (
	"errors"
)

func main() {
	errCmp := errors.New("xxx")
	println(5 == errCmp)
}

// Error:
// main/files/types/eql_0b4_stdlibs.gno:9:10: <untyped> bigint does not implement .uverse.error (missing method Error)
