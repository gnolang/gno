package main

import (
	"strconv"
)

type Error int8

func (e Error) Error() string {
	return "error: " + strconv.Itoa(int(e))
}

// left is untyped const, right is typed const
// untyped value can be converted to its correspondence value when:
// a) it's (untyped) const
// b) it's declared type of value, like type Error int8, 0 will be able to converted to Error
func main() {
	if 1 == Error(1) {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Output:
// what the firetruck?
