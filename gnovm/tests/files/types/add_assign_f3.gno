package main

import (
	"strconv"
)

type E interface {
	Error() string
}

type Error1 int64

func (e Error1) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

type Error2 int

func (e Error2) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

// both not const, and both interface
func main() {
	var e1 E = Error1(0)
	var e2 E = Error2(0)
	e1 += e2
	println(e1)
}

// Error:
// main/files/types/add_assign_f3.gno:27:2: operator += not defined on: InterfaceKind

// TypeCheckError:
// main/files/types/add_assign_f3.gno:27:2: invalid operation: operator + not defined on e1 (variable of interface type E)
