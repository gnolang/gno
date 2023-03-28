// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	ptests "gno.land/p/demo/tests"
	rtests "gno.land/r/demo/tests"
)

func lol() {
	println("HACK1", std.GetCaller())
	println("HACK2", rtests.GetCaller())
}

func main() {
	println(`[DEBUG]
	user1.gno:             g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
	crossrealm_test:       g1vla5mffzum6060t99u4xhm8mnhgxr0sz4k574p
	gno.land/r/demo/tests: g1gz4ycmx0s6ln2wdrsh4e00l9fsel2wskqa3snq
	gno.land/p/demo/tests: g1lc7c8nv62nqyyhhxe88tpxx786gwq68prx3f6e
	`)

	println("user1.gno -> gno.land/r/crossrealm_test: caller:", std.GetCaller())
	println("user1.gno -> gno.land/r/crossrealm_test -> gno.land/p/demo/tests: caller:", ptests.GetCaller())
	println("user1.gno -> gno.land/r/crossrealm_test -> gno.land/r/demo/tests: caller:", rtests.GetCaller()) // crossrealm -> gno.land/r/demo/tests

	rtests.Exec(func() {
		println("r/EXEC1", std.GetCaller())    // ??
		println("r/EXEC2", ptests.GetCaller()) // ??
	})
	rtests.Exec(lol)

	ptests.Exec(func() {
		println("p/EXEC1", std.GetCaller())    // ??
		println("p/EXEC2", ptests.GetCaller()) // ??
	})
	ptests.Exec(lol)

	println("3.c subtestsCaller", rtests.GetSubtestsCaller()) // gno.land/r/demo/tests -> gno.land/r/demo/tests/subtests
	rtests.ExecFromTest()
	println("4. main   :", std.GetCaller())
}

// Output:
// struct{("modified" string)}
