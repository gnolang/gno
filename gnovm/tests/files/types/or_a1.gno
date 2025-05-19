package main

import (
	"strconv"
)

type Error1 int64

func (e Error1) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

type Error2 int

func (e Error2) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

// both typed(different) const
func main() {
	println(Error1(0) | Error2(0))
}

// Error:
// main/or_a1.gno:21:10-31: invalid operation: (const (0 main.Error1)) | (const (0 main.Error2)) (mismatched types main.Error1 and main.Error2)

// TypeCheckError:
// main/or_a1.gno:21:10: invalid operation: Error1(0) | Error2(0) (mismatched types Error1 and Error2)
