package main

import (
	"errors"
)

func main() {
	errCmp := errors.New("xxx")
	println(5 == errCmp)
}

// Error:
// interface conversion: gnolang.Type is *gnolang.NativeType, not gnolang.PrimitiveType
