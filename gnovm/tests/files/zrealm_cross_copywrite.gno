// PKGPATH: gno.land/r/test/copywrite
package copywrite

// A copy of a stored struct made and written while another realm is current
// owns its nested struct: G.I is still unread when G is copied, and the write
// to tmp.I must leave G.I as it was.

import "gno.land/r/tests/vm/crossrealm"

type Inner struct{ X int }

type Outer struct{ I Inner }

var G = Outer{I: Inner{X: 1}}

func main(cur realm) {
	crossrealm.ExecCrossing(cross(cur), func() string {
		tmp := G
		tmp.I.X = 99
		return ""
	})
	println(G.I.X)
}

// Output:
// 1
