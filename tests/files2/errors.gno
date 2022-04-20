package main

import "errors"

func makeError() error {
	return errors.New("some error")
}

func main() {
	println(makeError().Error())
}

// Output:
// some error
