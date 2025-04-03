package main

import "github.com/gnolang/gno/_test/foo"

func main() { println(foo.Bar, foo.Boo) }

// Init functions of dependencies are executed separatedly from the test itself,
// so they don't print with the test proper.

// Output:
// BARR Boo
