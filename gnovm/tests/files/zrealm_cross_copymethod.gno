// PKGPATH: gno.land/r/test/copymethod
package copymethod

// A value method called on a stored struct while another realm is current
// runs on a copy of it: G.I is still unread when the receiver is copied, and
// the write to the copy's I must leave G.I as it was.

import "gno.land/r/tests/vm/crossrealm"

type Inner struct{ X int }

type Outer struct{ I Inner }

var G = Outer{I: Inner{X: 1}}

func (o Outer) Set(x int) { o.I.X = x }

func main(cur realm) {
	crossrealm.ExecCrossing(cross(cur), func() string {
		G.Set(77)
		return ""
	})
	println(G.I.X)
}

// Output:
// 1
