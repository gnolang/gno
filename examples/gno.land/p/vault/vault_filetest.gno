package main

import "gno.land/p/grc/grc20"
import "gno.land/p/vault"

func main() {
	fooToken := grc20.NewToken("Foo", "FOO", 4)
	vault := New(fooToken)

	println("TEST2", vault)
}

// Output:
// TEST2
