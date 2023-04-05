package main

import (
	"github.com/gnolang/gno/_test/bar"     // pkg name is actually quux
	baz "github.com/gnolang/gno/_test/baz" // pkg name is also quux, force it to be baz.
)

func main() {
	println("Hello", quux.Quux())
	println("Hello", baz.Quux())
}

// Output:
// Hello bar
// Hello baz
