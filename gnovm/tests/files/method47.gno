// https://github.com/gnolang/gno/issues/5957
package main

type D2 struct{}

func (D2) Foo() string { return "m" }

type D1 *D2 // defined pointer type: empty method set

func main() {
	var x D1
	_ = x.Foo // want: method not found, not "should not happen"
}

// Error:
// main/method47.gno:12:6-11: missing field Foo in main.D1

// TypeCheckError:
// main/method47.gno:12:8: x.Foo undefined (type D1 has no field or method Foo)
