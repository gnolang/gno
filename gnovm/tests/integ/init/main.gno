package main

var _ = func() int {
	println("HELLO HELLO!!")
	return 1
}()

func init() {
	println("HELLO WORLD!")
}
