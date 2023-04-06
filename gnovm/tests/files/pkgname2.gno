package main

import (
	"github.com/gnolang/gno/_test/toto" // pkg name is actually titi
)

func main() {
	println("Hello", titi.Quux())
}

// Output:
// Hello bar
