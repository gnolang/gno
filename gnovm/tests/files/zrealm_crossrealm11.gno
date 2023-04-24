// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	ptests "gno.land/p/demo/tests"
	"gno.land/p/demo/ufmt"
	rtests "gno.land/r/demo/tests"
	testfoo "gno.land/r/demo/tests_foo"
)

func Exec(fn func()) {
	fn()
}

func main() {
	// Create a map of the potential callers, this will give more understandable
	// output than the bech32 addresses.
	callersByAddr := make(map[std.Address]string)
	for _, caller := range []string{
		"user1.gno", "gno.land/r/crossrealm_test", "gno.land/r/demo/tests",
	} {
		addr := std.DerivePkgAddr(caller)
		callersByAddr[addr] = caller
	}

	baseCallStack := "user1.gno -> r/crossrealm_test.main()"
	tests := []struct {
		callStackAdd string
		callerFn     func() std.Address
	}{
		{
			callStackAdd: "",
			callerFn:     std.GetCaller,
		},
		{
			callStackAdd: " -> r/crossrealm_test.getCaller()",
			callerFn:     getCaller,
		},
		{
			callStackAdd: " -> p/demo/tests",
			callerFn:     ptests.GetCaller,
		},
		{
			callStackAdd: " -> p/demo/tests -> p/demo/tests/subtests",
			callerFn:     ptests.GetSubtestsCaller,
		},
		{
			callStackAdd: " -> r/demo/tests",
			callerFn:     rtests.GetCaller,
		},
		{
			callStackAdd: " -> r/demo/tests -> r/demo/tests/subtests",
			callerFn:     rtests.GetSubtestsCaller,
		},
		{
			callStackAdd: " -> p/demo/tests -> r/demo/tests",
			callerFn:     ptests.GetRealmGetCaller,
		},
	}
	for i, tt := range tests {
		callStack := ufmt.Sprintf("%s%s -> std.GetCaller()", baseCallStack, tt.callStackAdd)
		println(ufmt.Sprintf("STACK CASE %d:\n%s", i, callStack))
		println(ufmt.Sprintf("\tcaller is %s", callersByAddr[tt.callerFn()]))
		Exec(func() {
			println(ufmt.Sprintf("wrapped in r/crossrealm_test.Exec:\n\tcaller is %s", callersByAddr[tt.callerFn()]))
		})
		rtests.Exec(func() {
			println(ufmt.Sprintf("wrapped in r/demo/tests.Exec:\n\tcaller is %s", callersByAddr[tt.callerFn()]))
		})
		ptests.Exec(func() {
			println(ufmt.Sprintf("wrapped p/demo/tests.Exec:\n\tcaller is %s", callersByAddr[tt.callerFn()]))
		})
		println()
	}
}

func getCaller() std.Address {
	return std.GetCaller()
}

// Output:
// STACK CASE 0:
// user1.gno -> r/crossrealm_test.main() -> std.GetCaller()
// 	caller is user1.gno
// wrapped in r/crossrealm_test.Exec:
// 	caller is user1.gno
// wrapped in r/demo/tests.Exec:
// 	caller is gno.land/r/demo/tests
// wrapped p/demo/tests.Exec:
// 	caller is user1.gno
//
// STACK CASE 1:
// user1.gno -> r/crossrealm_test.main() -> r/crossrealm_test.getCaller() -> std.GetCaller()
// 	caller is user1.gno
// wrapped in r/crossrealm_test.Exec:
// 	caller is user1.gno
// wrapped in r/demo/tests.Exec:
// 	caller is gno.land/r/demo/tests
// wrapped p/demo/tests.Exec:
// 	caller is user1.gno
//
// STACK CASE 2:
// user1.gno -> r/crossrealm_test.main() -> p/demo/tests -> std.GetCaller()
// 	caller is user1.gno
// wrapped in r/crossrealm_test.Exec:
// 	caller is user1.gno
// wrapped in r/demo/tests.Exec:
// 	caller is gno.land/r/demo/tests
// wrapped p/demo/tests.Exec:
// 	caller is user1.gno
//
// STACK CASE 3:
// user1.gno -> r/crossrealm_test.main() -> p/demo/tests -> p/demo/tests/subtests -> std.GetCaller()
// 	caller is user1.gno
// wrapped in r/crossrealm_test.Exec:
// 	caller is user1.gno
// wrapped in r/demo/tests.Exec:
// 	caller is gno.land/r/demo/tests
// wrapped p/demo/tests.Exec:
// 	caller is user1.gno
//
// STACK CASE 4:
// user1.gno -> r/crossrealm_test.main() -> r/demo/tests -> std.GetCaller()
// 	caller is gno.land/r/crossrealm_test
// wrapped in r/crossrealm_test.Exec:
// 	caller is gno.land/r/crossrealm_test
// wrapped in r/demo/tests.Exec:
// 	caller is gno.land/r/crossrealm_test
// wrapped p/demo/tests.Exec:
// 	caller is gno.land/r/crossrealm_test
//
// STACK CASE 5:
// user1.gno -> r/crossrealm_test.main() -> r/demo/tests -> r/demo/tests/subtests -> std.GetCaller()
// 	caller is gno.land/r/demo/tests
// wrapped in r/crossrealm_test.Exec:
// 	caller is gno.land/r/demo/tests
// wrapped in r/demo/tests.Exec:
// 	caller is gno.land/r/demo/tests
// wrapped p/demo/tests.Exec:
// 	caller is gno.land/r/demo/tests
//
// STACK CASE 6:
// user1.gno -> r/crossrealm_test.main() -> p/demo/tests -> r/demo/tests -> std.GetCaller()
// 	caller is gno.land/r/crossrealm_test
// wrapped in r/crossrealm_test.Exec:
// 	caller is gno.land/r/crossrealm_test
// wrapped in r/demo/tests.Exec:
// 	caller is gno.land/r/crossrealm_test
// wrapped p/demo/tests.Exec:
// 	caller is gno.land/r/crossrealm_test
