package main

import "errors"

func makeError() error {
	return errors.New("some error")
}

func main() {
	var a error
	a = makeError()
	b := a.Error()
	_ = b // keep the variable used; the test is about something else
	println(a)
}

// Output:
// some error
