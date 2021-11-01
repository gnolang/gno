package gno

import (
	"fmt"
)

// NOTE: the golang compiler doesn't seem to be intelligent
// enough to remove steps when const debug is True,
// so it is still faster to first check the truth value
// before calling debug.Println or debug.Printf.

const debug debugging = false // or flip

type debugging bool

func (d debugging) Println(args ...interface{}) {
	if d {
		fmt.Println(append([]interface{}{"DEBUG:"}, args...)...)
	}
}

func (d debugging) Printf(format string, args ...interface{}) {
	if d {
		fmt.Printf("DEBUG: "+format, args...)
	}
}

var derrors []string = nil

// Instead of actually panic'ing, which messes with tests, errors are sometimes
// collected onto `var derrors`.  tests/file_test.go checks derrors after each
// test, and the file test fails if any unexpected debug errrors were found.
func (d debugging) Errorf(format string, args ...interface{}) {
	if d {
		derrors = append(derrors, fmt.Sprintf(format, args...))
	}
}

//----------------------------------------
// Exposed errors accessors
// File tests may access debug errors.

func HasDebugErrors() bool {
	return len(derrors) > 0
}

func GetDebugErrors() []string {
	return derrors
}

func ClearDebugErrors() {
	derrors = nil
}
