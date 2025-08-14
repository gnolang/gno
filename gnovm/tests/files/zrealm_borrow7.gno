// PKGPATH: gno.land/r/borrow
package borrow

import "gno.land/p/demo/tests/p_crossrealm"

func init() {
	p_crossrealm.F.Modify2()
}

func main() {
	println("F after modify: ", p_crossrealm.F)
	println("Err after modify: ", p_crossrealm.Err)
}

// Error:
// cannot directly modify readonly tainted object (w/o method): f<VPBlock(1,0)>.num
