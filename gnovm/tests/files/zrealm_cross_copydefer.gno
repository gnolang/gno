// PKGPATH: gno.land/r/test/copydefer
package copydefer

// A deferred call writing the named result of a function that returns a
// stored struct, run while another realm is current, writes the result, a
// copy: G.I is still unread at the return, and the write must leave G.I as
// it was.

import "gno.land/r/tests/vm/crossrealm"

type Inner struct{ X int }

type Outer struct{ I Inner }

var G = Outer{I: Inner{X: 1}}

func returned() (r Outer) {
	defer func() { r.I.X = 99 }()
	return G
}

func main(cur realm) {
	crossrealm.ExecCrossing(cross(cur), func() string {
		_ = returned()
		return ""
	})
	println(G.I.X)
}

// Output:
// 1
