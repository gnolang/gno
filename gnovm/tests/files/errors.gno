package main

import "errors"

func makeError() error {
	return errors.New("some error")
}

func main() {
	println(makeError())
}

// Output:
// some error
