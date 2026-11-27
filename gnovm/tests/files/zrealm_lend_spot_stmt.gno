// PKGPATH: gno.land/r/tests/vm/lendspotstmt
package lendspotstmt

var count int

func bump() {
	func() { count++ }()
	func() {
		f := func() { count++ }
		f()
	}()
	// Deferred, it runs before bump returns, in bump's own mode: accepted.
	defer func() { count++ }()
}

// A literal called where it is written runs in the mode of the code around
// it: in a crossing call it writes.
func main(cur realm) {
	bump()
	println("s03 statement, helper inside, deferred", count)
}

// Output:
// s03 statement, helper inside, deferred 3
