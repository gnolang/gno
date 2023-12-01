package main

import (
	"errors"
	"strconv"
)

type Error int64

func (e Error) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

var errCmp = errors.New("XXXX")

// NOTE: in golang, this won't compile by type mismatch
// code like this : won't compile by : "invalid operation: operator + not defined on err1 (variable of type error)"
// which indicates that golang will check type match first, and operand legality secondly
// that is, for op =+, first check type match, second check operand
// in our condition, should check convert first, after that, check operand?

//package main
//
//import (
//"errors"
//)
//
//var err1 = errors.New("XXXX")
//var err2 = errors.New("XXXX")
//
//func main() {
//	err1 += err2
//	println(err1)
//}

// specil case:
// one is interface
func main() {
	r := Error(0)
	errCmp += r
	println(errCmp)
}

// Error:
// main/files/types/13f2_filetest.gno:40: operator += not defined on: .uverse.error
