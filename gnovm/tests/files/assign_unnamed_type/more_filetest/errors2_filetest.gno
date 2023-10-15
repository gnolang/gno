package main

import "errors"

func makeError() error {
	return errors.New("some error")
}

func main() {
	var a error
	a = makeError()
	b := a.Error()
	println(a)
}

// Output:
// some error
