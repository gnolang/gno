// https://github.com/gnolang/gno/issues/5957 (review follow-up)
// The (*x).f shorthand does not apply when the pointed-to type is
// itself a (defined) pointer type: *D1 exposes nothing.
package main

type D2 struct{ A int }

type D1 *D2

func main() {
	var p *D1
	_ = p.A
}

// Error:
// main/ptr14.gno:12:6-9: missing field A in *main.D1

// TypeCheckError:
// main/ptr14.gno:12:8: p.A undefined (type *D1 has no field or method A)
