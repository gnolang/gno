package main

import (
	"errors"
)

func main() {
	if func() bool {
		println(errors.New("imported").Error())
		return false
	}() {
		return
	}
}

// Output:
// imported
