package main

import (
	"errors"
)

func main() {
	errCmp := errors.New("xxx")
	println(5 == errCmp)
}

// Error:
// main/files/types/0b4_stdlibs_filetest.gno:9: <untyped> bigint does not implement .uverse.error
