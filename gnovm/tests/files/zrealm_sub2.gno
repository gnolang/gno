// PKGPATH: gno.land/r/demo/subrealm
package subrealm

import "chain"

// Off-chain-computable derivation: chain.DerivePkgSubAddr matches
// Sub().Address(), and chain.SplitPkgSubPath recovers (host, subpath).
func main(cur realm) {
	sub := cur.Sub("dao/42")
	println(sub.Address() == chain.DerivePkgSubAddr("gno.land/r/demo/subrealm", "dao/42"))
	host, subpath, ok := chain.SplitPkgSubPath(sub.PkgPath())
	println(host, subpath, ok)
	host, subpath, ok = chain.SplitPkgSubPath(cur.PkgPath())
	println(host, subpath, ok)
}

// Output:
// true
// gno.land/r/demo/subrealm dao/42 true
// gno.land/r/demo/subrealm  false
