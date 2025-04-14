package main

// These imports require an identifier, as the `package` declaration doesn't
// match their paths.

import (
	tutu "github.com/gnolang/gno/_test/tata"
	titi "github.com/gnolang/gno/_test/toto"
)

func main() {
	println("Hello", titi.Quux())
	println("Hello", tutu.Quux())
}

// Output:
// Hello bar
// Hello baz
