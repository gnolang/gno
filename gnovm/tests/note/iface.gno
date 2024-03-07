package main

type MyError struct {
	error
}

var err error

func main() {
	err = (*MyError)(nil)
	println(err)
}

// Output:
