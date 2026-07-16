// https://github.com/gnolang/gno/issues/5957
// A pointer to a defined interface type exposes no methods.
package main

type B interface{ M() string }
type A *B

func main() {
	var x A
	_ = x.M
}

// Error:
// main/ptr13.gno:10:6-9: missing field M in main.A

// TypeCheckError:
// main/ptr13.gno:10:8: x.M undefined (type A is pointer to interface, not interface)
