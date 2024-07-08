package main

import (
	"github.com/gnolang/gno/_test/tata" // pkg name is actually tutu
	"github.com/gnolang/gno/_test/toto" // pkg name is actually titi
)

func main() {
	println("Hello", titi.Quux())
	println("Hello", tutu.Quux())
}

// Output:
// Hello bar
// Hello baz
