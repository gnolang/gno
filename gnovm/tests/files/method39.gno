// This test with method38.gno are answers to this issue: https://github.com/gnolang/gno/issues/4789
package main

type Base struct{}

func (Base) Foo() {}

type Derived Base // declared type does not inherit methods

func main() {
	var d Derived
	d.Foo() // want: method not found
}

// Error:
// main/method39.gno:12:2-7: missing field Foo in main.Derived

// TypeCheckError:
// main/method39.gno:12:4: d.Foo undefined (type Derived has no field or method Foo)
