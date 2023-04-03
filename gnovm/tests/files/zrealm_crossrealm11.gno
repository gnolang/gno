// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	ptests "gno.land/p/demo/tests"
	rtests "gno.land/r/demo/tests"
)

func Exec(fn func()) {
	fn()
} 

func main() {
	println(`[DEBUG]
	user1.gno:             g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
	crossrealm_test:       g1vla5mffzum6060t99u4xhm8mnhgxr0sz4k574p
	gno.land/p/demo/tests: g1lc7c8nv62nqyyhhxe88tpxx786gwq68prx3f6e
	gno.land/r/demo/tests: g1gz4ycmx0s6ln2wdrsh4e00l9fsel2wskqa3snq
	`)

	println("user1.gno -> r/crossrealm_test:", std.GetCaller())

	println("user1.gno -> r/crossrealm_test -> p/demo/tests:", ptests.GetCaller())

	println("user1.gno -> r/crossrealm_test -> r/demo/tests:", rtests.GetCaller())

	println("user1.gno -> r/crossrealm_test -> r/demo/tests -> r/demo/tests/subtests", rtests.GetSubtestsCaller())

	println("user1.gno -> r/crossrealm_test -> p/demo/tests -> p/demo/tests/subtests", ptests.GetSubtestsCaller())

	// let's try to exploit an unsafe realm
	rtests.Exec(func() {
		// caller is `r/demo/tests`
		println("user1.gno -> r/crossrealm_test -> r/demo/tests.Exec", std.GetCaller())

		
		// NOTE(WARNING): this one is a weird behaviour, because the caller returned is crossrealm_test
		// caller is `crossrealm_test`
		println("/!\\ user1.gno -> r/crossrealm_test -> r/demo/tests.Exec -> p/demo/tests:", ptests.GetCaller())
	})

	ptests.Exec(func() {
		println("user1.gno -> r/crossrealm_test -> p/demo/tests.Exec: ", std.GetCaller())
		println("user1.gno -> r/crossrealm_test -> p/demo/tests.Exec -> p/demo/tests:", ptests.GetCaller())
	})

	rtests.ExecFromTest()

	Exec(func() {
		println("user1.gno -> r/crossrealm_test.Exec:", std.GetCaller())
	})
}

// Output:
// [DEBUG]
// 	user1.gno:             g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// 	crossrealm_test:       g1vla5mffzum6060t99u4xhm8mnhgxr0sz4k574p
// 	gno.land/p/demo/tests: g1lc7c8nv62nqyyhhxe88tpxx786gwq68prx3f6e
// 	gno.land/r/demo/tests: g1gz4ycmx0s6ln2wdrsh4e00l9fsel2wskqa3snq
//
// user1.gno -> r/crossrealm_test: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// user1.gno -> r/crossrealm_test -> p/demo/tests: g1vla5mffzum6060t99u4xhm8mnhgxr0sz4k574p
// user1.gno -> r/crossrealm_test -> r/demo/tests: g1vla5mffzum6060t99u4xhm8mnhgxr0sz4k574p
// user1.gno -> r/crossrealm_test -> r/demo/tests -> r/demo/tests/subtests g1gz4ycmx0s6ln2wdrsh4e00l9fsel2wskqa3snq
// user1.gno -> r/crossrealm_test -> p/demo/tests -> p/demo/tests/subtests g1vla5mffzum6060t99u4xhm8mnhgxr0sz4k574p
// user1.gno -> r/crossrealm_test -> r/demo/tests.Exec g1gz4ycmx0s6ln2wdrsh4e00l9fsel2wskqa3snq
// /!\ user1.gno -> r/crossrealm_test -> r/demo/tests.Exec -> p/demo/tests: g1vla5mffzum6060t99u4xhm8mnhgxr0sz4k574p
// user1.gno -> r/crossrealm_test -> p/demo/tests.Exec:  g1vla5mffzum6060t99u4xhm8mnhgxr0sz4k574p
// user1.gno -> r/crossrealm_test -> p/demo/tests.Exec -> p/demo/tests: g1vla5mffzum6060t99u4xhm8mnhgxr0sz4k574p
// xxx -> r/demo/tests -> r/demo/tests.Exec: g1gz4ycmx0s6ln2wdrsh4e00l9fsel2wskqa3snq
// user1.gno -> r/crossrealm_test.Exec: g1vla5mffzum6060t99u4xhm8mnhgxr0sz4k574p
