// This test with method39.gno are answers to this issue: https://github.com/gnolang/gno/issues/4789
package main

import "fmt"

type B struct{}

func (B) Foo() string { return "Foo from B" }

type A struct{ B } // embeds B
type T A           // T retains the struct layout of A (embedding B)
type S struct{ T } // S embeds T (which embeds B)

func main() {
	var s S
	fmt.Println(s.Foo())
	fmt.Println("ok")
}

// Output:
// Foo from B
// ok
