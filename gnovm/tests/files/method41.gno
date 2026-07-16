// https://github.com/gnolang/gno/issues/5957
package main

type E struct{}

func (E) Bar() string { return "bar" }

type D2 struct{ E }

type D1 *D2 // defined pointer type: no methods promote through it

func main() {
	var x D1
	_ = x.Bar // want: method not found (promoted via embedded E in *D2)
}

// Error:
// main/method41.gno:14:6-11: missing field Bar in main.D1

// TypeCheckError:
// main/method41.gno:14:8: x.Bar undefined (type D1 has no field or method Bar)
