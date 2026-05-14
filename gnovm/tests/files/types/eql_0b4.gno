package main

import (
	"errors"
)

func main() {
	errCmp := errors.New("xxx")
	println(5 == errCmp)
}

// Error:
// main/eql_0b4.gno:9:10-21: invalid operation: (mismatched types <untyped> bigint and .uverse.error)

// TypeCheckError:
// main/eql_0b4.gno:9:10: cannot convert 5 (untyped int constant) to type interface{Error() string}
