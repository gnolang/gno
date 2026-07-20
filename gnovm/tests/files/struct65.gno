// https://github.com/gnolang/gno/issues/5957 (review follow-up)
package main

type I interface{ M() string }

type S struct{ *I } // want: embedded field type cannot be a pointer to an interface

func main() {
	var s S
	_ = s.M
}

// Error:
// main/struct65.gno:6:8-20: embedded field type cannot be a pointer to an interface: *main.I

// TypeCheckError:
// main/struct65.gno:6:16: embedded field type cannot be a pointer to an interface
