// https://github.com/gnolang/gno/issues/5957
package main

type D2 struct{}

func (D2) Foo() string { return "m" }

type D1 *D2

type S struct{ D1 } // want: embedded field type cannot be a pointer

func main() {
	var s S
	_ = s.Foo
}

// Error:
// main/struct64.gno:10:8-20: embedded field type cannot be a pointer: main.D1

// TypeCheckError:
// main/struct64.gno:10:16: embedded field type cannot be a pointer; main/struct64.gno:14:8: s.Foo undefined (type S has no field or method Foo)
