// PKGPATH: gno.land/r/test
package test

var root interface{}

func main() {
	println(root)
	root = 1
	println(root)
}

// Output:
// nil
// 1

// Realm:
// u[{{A8ADA09DEE16D791FD406D629FE29BB0ED084A30} 0}]=ValuePreimage{4:0FCED9A7AC6294E7FF7C3E571D1B7FB49A33A8BF:0,0,0}

// NOTE:
// the above is showing that the realm's block (of 1 variable) changed.
// TODO: show preimage for hash calculation.
