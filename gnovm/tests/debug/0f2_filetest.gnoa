package main

import (
	"errors"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

var errCmp = errors.New("XXXX")

// specil case:
// one is interface
// TODO: seems what we can do is to not allowed gonative non-empty interface come into gno
// namely, we don't run this in .native mode, (withNativeLibs)
func main() {
	if Error("XXXX") == errCmp { // this would not convert, errCmp is interface, refer to preprocess, line2526
		//if errCmp == 1 {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Output:
// something else
