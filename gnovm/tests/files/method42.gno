// https://github.com/gnolang/gno/issues/5957
package main

type I interface{ Baz() string }

type D2 struct{ I }

type D1 *D2 // defined pointer type: interface methods don't promote either

func main() {
	var x D1
	_ = x.Baz
}

// Error:
// main/method42.gno:12:6-11: missing field Baz in main.D1

// TypeCheckError:
// main/method42.gno:12:8: x.Baz undefined (type D1 has no field or method Baz)
