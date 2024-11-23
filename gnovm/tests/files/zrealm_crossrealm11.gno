// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	ptests "gno.land/p/demo/tests"
	"gno.land/p/demo/ufmt"
	rtests "gno.land/r/demo/tests"
	"std"
)

func getPrevRealm() std.Realm {
	return std.PrevRealm()
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

	assertRealm := func(r std.Realm) {
		pkgPath := callersByAddr[r.Addr()]
		if r.IsUser() && pkgPath != "user1.gno" {
			panic(ufmt.Sprintf("ERROR: expected: 'user1.gno', got:'%s'", pkgPath))
		} else if !r.IsUser() && pkgPath != r.PkgPath() {
			panic(ufmt.Sprintf("ERROR: expected: '%s', got: '%s'", pkgPath, r.PkgPath()))
		}
	}

	tests := []struct {
		callStackAdd string
		callerFn     func() std.Realm
	}{
		{
			callStackAdd: "",
			callerFn:     std.PrevRealm,
		},
		{
			callStackAdd: " -> r/crossrealm_test.getPrevRealm",
			callerFn:     getPrevRealm,
		},
		{
			callStackAdd: " -> p/demo/tests",
			callerFn:     ptests.GetPrevRealm,
		},
		{
			callStackAdd: " -> p/demo/tests -> p/demo/tests/subtests",
			callerFn:     ptests.GetPSubtestsPrevRealm,
		},
		{
			callStackAdd: " -> r/demo/tests",
			callerFn:     rtests.GetPrevRealm,
		},
		{
			callStackAdd: " -> r/demo/tests -> r/demo/tests/subtests",
			callerFn:     rtests.GetRSubtestsPrevRealm,
		},
		{
			callStackAdd: " -> p/demo/tests -> r/demo/tests",
			callerFn:     ptests.GetRTestsGetPrevRealm,
		},
	}

	println("---") // needed to have space prefixes
	printColumns("STACK", "std.PrevRealm")
	printColumns("-----", "------------------")

	baseCallStack := "user1.gno -> r/crossrealm_test.main"
	for i, tt := range tests {
		printColumns(baseCallStack+tt.callStackAdd, callersByAddr[tt.callerFn().Addr()])
		Exec(func() {
			r := tt.callerFn()
			assertRealm(r)
			printColumns(baseCallStack+" -> r/crossrealm_test.Exec"+tt.callStackAdd, callersByAddr[r.Addr()])
		})
		rtests.Exec(func() {
			r := tt.callerFn()
			assertRealm(r)
			printColumns(baseCallStack+" -> r/demo/tests.Exec"+tt.callStackAdd, callersByAddr[r.Addr()])
		})
		ptests.Exec(func() {
			r := tt.callerFn()
			assertRealm(r)
			printColumns(baseCallStack+" -> p/demo/tests.Exec"+tt.callStackAdd, callersByAddr[r.Addr()])
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
//                                                                                                     STACK = std.PrevRealm
//                                                                                                     ----- = ------------------
//                                                                       user1.gno -> r/crossrealm_test.main = user1.gno
//                                             user1.gno -> r/crossrealm_test.main -> r/crossrealm_test.Exec = user1.gno
//                                                  user1.gno -> r/crossrealm_test.main -> r/demo/tests.Exec = gno.land/r/demo/tests
//                                                  user1.gno -> r/crossrealm_test.main -> p/demo/tests.Exec = user1.gno
//                                     user1.gno -> r/crossrealm_test.main -> r/crossrealm_test.getPrevRealm = user1.gno
//           user1.gno -> r/crossrealm_test.main -> r/crossrealm_test.Exec -> r/crossrealm_test.getPrevRealm = user1.gno
//                user1.gno -> r/crossrealm_test.main -> r/demo/tests.Exec -> r/crossrealm_test.getPrevRealm = gno.land/r/demo/tests
//                user1.gno -> r/crossrealm_test.main -> p/demo/tests.Exec -> r/crossrealm_test.getPrevRealm = user1.gno
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
