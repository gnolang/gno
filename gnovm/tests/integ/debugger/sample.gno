// This is a sample target gno program to test the gnovm debugger.
// See ../../cmd/gno/debug_test.go for the debugger test cases.

package main

func f(name string, i int) {
	println(name, i)
}

func g(s string, n int) {
	f(s, n)
}

func main() {
	println("before f")
	g("hello", 3)
	println("bye")
}
