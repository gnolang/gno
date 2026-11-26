// PKGPATH: gno.land/r/zrealm_chalc_testing_exempt
package zrealm_chalc_testing_exempt

import "testing"

var count int

func attack() { run(func() { count++ }) }

// The MEASUREMENT exemption: a callback run by the test-only testing package.
func run(g func()) { testing.AllocsPerRun2(1, g) }

func main(cur realm) {
	println("before:", count)
	attack()
	println("after:", count)
}

// Output:
// before: 0
// after: 1
