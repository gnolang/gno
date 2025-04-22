package main

import (
	"errors"
)

func main() {
	errCmp := errors.New("xxx")
	println(5 == errCmp)
}

// Error:
// main/files/types/eql_0b4.gno:9:10: <untyped> bigint does not implement .uverse.error (missing method Error)

// TypeCheckError:
// main/files/types/eql_0b4.gno:9:10: cannot convert 5 (untyped int constant) to type interface{Error() string}
