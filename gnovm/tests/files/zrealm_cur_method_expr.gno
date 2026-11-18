// PKGPATH: gno.land/r/demo/curmethexpr
package curmethexpr

// A METHOD EXPRESSION passes the receiver as an ordinary first argument, so the
// frame carries no receiver while the block still starts with one. `cur` is
// therefore at block index 1, as it is for a bound call — keying that off
// fr.Receiver instead of the declaration read the RECEIVER as the cur and
// rejected this legitimate call, since a *S is not a realm and matches no
// caller identity.
//
// Invisible until crossing methods started entering the check at all: before
// that the whole block was skipped for methods, so the mis-indexing was a
// silent no-op rather than a panic.

type S struct{ tag string }

func (s *S) m(cur realm) {
	println("m:", s.tag, cur.PkgPath(), cur.IsCurrent())
}

func main(cur realm) {
	s := &S{tag: "real"}
	s.m(cur)    // bound: receiver on the frame
	f := (*S).m // method expression: receiver as an argument
	f(s, cur)
	g := s.m // method value: receiver bound into the closure
	g(cur)
}

// Output:
// m: real gno.land/r/demo/curmethexpr true
// m: real gno.land/r/demo/curmethexpr true
// m: real gno.land/r/demo/curmethexpr true
