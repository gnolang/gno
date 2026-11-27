package main

import "gno.land/p/demo/tests/recoverfv"

// The same through a library's method from package main.
func main() {
	c := recoverfv.New()
	ops := []func(){func() { c.Set("x") }}
	rec := func() (rec any) {
		defer func() { rec = recover() }()
		ops[0]()
		return
	}()
	println("via value:", rec)
	println("OUTCOME caught")
}

// Output:
// via value: int is not of type string
// OUTCOME caught
