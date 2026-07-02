// PKGPATH: gno.land/r/demo/subrealm
package subrealm

import "chain"

// Basic Sub() minting: synthesized pkgpath, derived address,
// classification methods, IsCurrent relaxation.
func main(cur realm) {
	sub := cur.Sub("dao/42")
	println(sub.PkgPath())
	println(sub.Address())
	println(sub.Address() == chain.PackageAddress("gno.land/r/demo/subrealm:dao/42"))
	println(sub.IsCurrent())
	println(sub.IsCode(), sub.IsUser(), sub.IsUserCall(), sub.IsUserRun(), sub.IsEphemeral())
	println(sub.String())
}

// Output:
// gno.land/r/demo/subrealm:dao/42
// g1mu3u0pzc6td7u5jqphm6ecacjkxrmftfy4hl4e
// true
// true
// true false false false false
// realm{gno.land/r/demo/subrealm:dao/42:g1mu3u0pzc6td7u5jqphm6ecacjkxrmftfy4hl4e}
