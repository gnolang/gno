// PKGPATH: gno.land/r/aliascross
package aliascross

import "gno.land/p/nt/uassert/v0"

// type alias to another user package's DeclaredType (an interface).
type TestingT = uassert.TestingT

func main(cur realm) {
	println("ok")
}

// Output:
// ok

// Types:
// TestingT[gno.land/p/nt/uassert/v0.TestingT]=
//     {
//         "ID": "gno.land/p/nt/uassert/v0.TestingT"
//     }
