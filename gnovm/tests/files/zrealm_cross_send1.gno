// PKGPATH: gno.land/r/demo/crosssend1
package crosssend1

func Deposit(cur realm) {}

func main(cur realm) {
	Deposit(cross(cur, 5))
}

// Error:
// gno.land/r/demo/crosssend1/zrealm_cross_send1.gno:7:10-23: the second argument of cross must be a chain.Coins value
