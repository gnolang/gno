// The operand satisfies the interface via an embedded interface FIELD, so
// call-time resolution takes a nested lazy hop whose lookup must stay
// qualified by the bind site (MethodPkg propagates per hop). Mutating the
// field between bind and call re-dispatches to the new concrete type,
// still under the bind-site qualifier. Verified against real Go.
package main

import "filetests/extern/secx"

type Box struct {
	secx.Sec
}

func main() {
	b := &Box{Sec: secx.SecImpl{}}
	gd := secx.MakeGetter(b)
	println("D:", gd())

	b.Sec = secx.SecImpl2{}
	println("E:", gd())
}

// Output:
// D: 7
// E: 8
