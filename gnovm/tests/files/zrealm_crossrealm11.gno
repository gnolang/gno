// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"
	"strings"

	ptests "gno.land/p/demo/tests"
	"gno.land/p/demo/ufmt"
	rtests "gno.land/r/demo/tests"
	testfoo "gno.land/r/demo/tests_foo"
)

func getRealmCaller() std.Address {
	return std.GetRealmCaller()
}

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

	tests := []struct {
		callStackAdd string
		callerFn     func() std.Address
	}{
		{
			callStackAdd: "",
			callerFn:     std.GetRealmCaller,
		},
		{
			callStackAdd: " -> r/crossrealm_test.getRealmCaller",
			callerFn:     getRealmCaller,
		},
		{
			callStackAdd: " -> p/demo/tests",
			callerFn:     ptests.GetRealmCaller,
		},
		{
			callStackAdd: " -> p/demo/tests -> p/demo/tests/subtests",
			callerFn:     ptests.GetPSubtestsCaller,
		},
		{
			callStackAdd: " -> r/demo/tests",
			callerFn:     rtests.GetRealmCaller,
		},
		{
			callStackAdd: " -> r/demo/tests -> r/demo/tests/subtests",
			callerFn:     rtests.GetPSubtestsCaller,
		},
		{
			callStackAdd: " -> p/demo/tests -> r/demo/tests",
			callerFn:     ptests.GetRTestsGetRealmCaller,
		},
	}

	println("---") // needed to have space prefixes
	printColumns("STACK", "std.GetRealmCaller")
	printColumns("-----", "------------------")

	baseCallStack := "user1.gno -> r/crossrealm_test.main"
	for i, tt := range tests {
		printColumns(baseCallStack+tt.callStackAdd, callersByAddr[tt.callerFn()])
		Exec(func() {
			printColumns(baseCallStack+" -> r/crossrealm_test.Exec"+tt.callStackAdd, callersByAddr[tt.callerFn()])
		})
		rtests.Exec(func() {
			printColumns(baseCallStack+" -> r/demo/tests.Exec"+tt.callStackAdd, callersByAddr[tt.callerFn()])
		})
		ptests.Exec(func() {
			printColumns(baseCallStack+" -> p/demo/tests.Exec"+tt.callStackAdd, callersByAddr[tt.callerFn()])
		})
	}
}

func printColumns(left, right string) {
	const w = 105

	output := ""
	padding := w - len(left)

	// strings.Repeat is not always available when using various imports modes.
	for i := 0; i < padding; i++ {
		output += " "
	}

	output += left
	output += " = "
	output += right
	println(output)
}

// Output:
// ---
//                                                                                                     STACK = std.GetRealmCaller
//                                                                                                     ----- = ------------------
//                                                                       user1.gno -> r/crossrealm_test.main = user1.gno
//                                             user1.gno -> r/crossrealm_test.main -> r/crossrealm_test.Exec = user1.gno
//                                                  user1.gno -> r/crossrealm_test.main -> r/demo/tests.Exec = gno.land/r/demo/tests
//                                                  user1.gno -> r/crossrealm_test.main -> p/demo/tests.Exec = user1.gno
//                                   user1.gno -> r/crossrealm_test.main -> r/crossrealm_test.getRealmCaller = user1.gno
//         user1.gno -> r/crossrealm_test.main -> r/crossrealm_test.Exec -> r/crossrealm_test.getRealmCaller = user1.gno
//              user1.gno -> r/crossrealm_test.main -> r/demo/tests.Exec -> r/crossrealm_test.getRealmCaller = gno.land/r/demo/tests
//              user1.gno -> r/crossrealm_test.main -> p/demo/tests.Exec -> r/crossrealm_test.getRealmCaller = user1.gno
//                                                       user1.gno -> r/crossrealm_test.main -> p/demo/tests = user1.gno
//                             user1.gno -> r/crossrealm_test.main -> r/crossrealm_test.Exec -> p/demo/tests = user1.gno
//                                  user1.gno -> r/crossrealm_test.main -> r/demo/tests.Exec -> p/demo/tests = gno.land/r/demo/tests
//                                  user1.gno -> r/crossrealm_test.main -> p/demo/tests.Exec -> p/demo/tests = user1.gno
//                              user1.gno -> r/crossrealm_test.main -> p/demo/tests -> p/demo/tests/subtests = user1.gno
//    user1.gno -> r/crossrealm_test.main -> r/crossrealm_test.Exec -> p/demo/tests -> p/demo/tests/subtests = user1.gno
//         user1.gno -> r/crossrealm_test.main -> r/demo/tests.Exec -> p/demo/tests -> p/demo/tests/subtests = gno.land/r/demo/tests
//         user1.gno -> r/crossrealm_test.main -> p/demo/tests.Exec -> p/demo/tests -> p/demo/tests/subtests = user1.gno
//                                                       user1.gno -> r/crossrealm_test.main -> r/demo/tests = gno.land/r/crossrealm_test
//                             user1.gno -> r/crossrealm_test.main -> r/crossrealm_test.Exec -> r/demo/tests = gno.land/r/crossrealm_test
//                                  user1.gno -> r/crossrealm_test.main -> r/demo/tests.Exec -> r/demo/tests = gno.land/r/crossrealm_test
//                                  user1.gno -> r/crossrealm_test.main -> p/demo/tests.Exec -> r/demo/tests = gno.land/r/crossrealm_test
//                              user1.gno -> r/crossrealm_test.main -> r/demo/tests -> r/demo/tests/subtests = gno.land/r/demo/tests
//    user1.gno -> r/crossrealm_test.main -> r/crossrealm_test.Exec -> r/demo/tests -> r/demo/tests/subtests = gno.land/r/demo/tests
//         user1.gno -> r/crossrealm_test.main -> r/demo/tests.Exec -> r/demo/tests -> r/demo/tests/subtests = gno.land/r/demo/tests
//         user1.gno -> r/crossrealm_test.main -> p/demo/tests.Exec -> r/demo/tests -> r/demo/tests/subtests = gno.land/r/demo/tests
//                                       user1.gno -> r/crossrealm_test.main -> p/demo/tests -> r/demo/tests = gno.land/r/crossrealm_test
//             user1.gno -> r/crossrealm_test.main -> r/crossrealm_test.Exec -> p/demo/tests -> r/demo/tests = gno.land/r/crossrealm_test
//                  user1.gno -> r/crossrealm_test.main -> r/demo/tests.Exec -> p/demo/tests -> r/demo/tests = gno.land/r/crossrealm_test
//                  user1.gno -> r/crossrealm_test.main -> p/demo/tests.Exec -> p/demo/tests -> r/demo/tests = gno.land/r/crossrealm_test
