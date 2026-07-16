// https://github.com/gnolang/gno/issues/5957
// Fields don't promote through two levels of defined pointer types:
// the (*x).f shorthand applies once, and (*x) is again a defined
// pointer type (B), which has no field F.
package main

func main() {
	type C struct{ F int }
	type B *C
	type A *B
	var x A
	_ = x.F
}

// Error:
// main/ptr12.gno:12:6-9: missing field F in main[main/ptr12.gno:7:1-13:2].A

// TypeCheckError:
// main/ptr12.gno:12:8: x.F undefined (type A has no field or method F)
