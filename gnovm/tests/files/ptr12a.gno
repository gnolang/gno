// run

// Test that new() and &var on zero-sized types return the same address
// (zerobase), including empty struct and nested zero-sized struct.
//
// Note: Go's gc compiler constant-folds &x==&y to false for distinct
// zero-sized variables (non-portable SSA optimization). Gno does not
// replicate that. Per the spec: "pointers to distinct zero-size
// variables may or may not be equal" — Gno chooses "may": same
// address ⇒ equal. Both outcomes are explicitly permitted.

package main

var e struct{}
var pe, qe = new(struct{}), new(struct{})

type nestedZero struct{ _ [0]int }
var n nestedZero
var pn, qn = new(nestedZero), new(nestedZero)

func main() {
	// empty struct
	if pe != qe {
		print("pe=", pe, " qe=", qe, " pe==qe = ", pe == qe, "\n")
		panic("FAIL")
	}
	if &e != pe {
		print("&e=", &e, " pe=", pe, " &e==pe = ", &e == pe, "\n")
		panic("FAIL")
	}

	// nested zero-sized struct: struct{ _ [0]int }
	if pn != qn {
		print("pn=", pn, " qn=", qn, " pn==qn = ", pn == qn, "\n")
		panic("FAIL")
	}
	if &n != pn {
		print("&n=", &n, " pn=", pn, " &n==pn = ", &n == pn, "\n")
		panic("FAIL")
	}

	println("ok")
}

// Output:
// ok
