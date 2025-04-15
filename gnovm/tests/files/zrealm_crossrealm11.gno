// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	ptests "gno.land/p/demo/tests"
	"gno.land/p/demo/ufmt"
	rtests "gno.land/r/demo/tests"
)

func getPreviousRealm() std.Realm {
	return std.PreviousRealm()
}

func Exec(fn func()) {
	fn()
}

func main() {
	crossing()

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
		pkgPath := callersByAddr[r.Address()]
		if r.IsUser() && pkgPath != "user1.gno" {
			panic(ufmt.Sprintf("ERROR: expected: 'user1.gno', got:'%s'", pkgPath))
		} else if !r.IsUser() && pkgPath != r.PkgPath() {
			panic(ufmt.Sprintf("ERROR: expected: '%s', got: '%s'", pkgPath, r.PkgPath()))
		}
	}

	tests := []struct {
		callStackAdd string
		callerFn     func() std.Realm
		isSwitch     bool
	}{
		{
			callStackAdd: " -> std.PreviousRealm",
			callerFn:     std.PreviousRealm,
			isSwitch:     false,
		},
		{
			callStackAdd: " -> r/crossrealm_test.getPreviousRealm",
			callerFn:     getPreviousRealm,
			isSwitch:     false,
		},
		{
			callStackAdd: " -> p/demo/tests",
			callerFn:     ptests.GetPreviousRealm,
			isSwitch:     false,
		},
		{
			callStackAdd: " -> p/demo/tests -> p/demo/tests/subtests",
			callerFn:     ptests.GetPSubtestsPreviousRealm,
			isSwitch:     false,
		},
		{
			callStackAdd: " !> r/demo/tests",
			callerFn:     rtests.GetPreviousRealm,
			isSwitch:     true,
		},
		{
			callStackAdd: " !> r/demo/tests !> r/demo/tests/subtests",
			callerFn:     rtests.GetRSubtestsPreviousRealm,
			isSwitch:     true,
		},
	}

	println("---") // needed to have space prefixes
	printColumns("STACK", "std.PreviousRealm")
	printColumns("-----", "------------------")

	baseCallStack := "user1.gno !> r/crossrealm_test.main2"
	for i, tt := range tests {
		{ // with no Exec
			var r std.Realm
			if tt.isSwitch {
				r = cross(tt.callerFn)()
			} else {
				r = tt.callerFn()
			}
			printColumns(baseCallStack+tt.callStackAdd, callersByAddr[r.Address()])

		}
		Exec(func() {
			var r std.Realm
			if tt.isSwitch {
				r = cross(tt.callerFn)()
			} else {
				r = tt.callerFn()
			}
			assertRealm(r)
			printColumns(baseCallStack+" -> r/crossrealm_test.Exec"+tt.callStackAdd, callersByAddr[r.Address()])
		})
		rtests.Exec(func() {
			var r std.Realm
			if tt.isSwitch {
				r = cross(tt.callerFn)()
			} else {
				r = tt.callerFn()
			}
			assertRealm(r)
			printColumns(baseCallStack+" -> r/demo/tests.Exec"+tt.callStackAdd, callersByAddr[r.Address()])
		})
		cross(rtests.ExecSwitch)(func() {
			var r std.Realm
			if tt.isSwitch {
				r = cross(tt.callerFn)()
			} else {
				r = tt.callerFn()
			}
			assertRealm(r)
			printColumns(baseCallStack+" !> r/demo/tests.ExecSwitch"+tt.callStackAdd, callersByAddr[r.Address()])
		})
		ptests.Exec(func() {
			var r std.Realm
			if tt.isSwitch {
				r = cross(tt.callerFn)()
			} else {
				r = tt.callerFn()
			}
			assertRealm(r)
			printColumns(baseCallStack+" -> p/demo/tests.Exec"+tt.callStackAdd, callersByAddr[r.Address()])
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
//                                                                                                     STACK = std.PreviousRealm
//                                                                                                     ----- = ------------------
//                                                 user1.gno !> r/crossrealm_test.main2 -> std.PreviousRealm = user1.gno
//                       user1.gno !> r/crossrealm_test.main2 -> r/crossrealm_test.Exec -> std.PreviousRealm = user1.gno
//                            user1.gno !> r/crossrealm_test.main2 -> r/demo/tests.Exec -> std.PreviousRealm = user1.gno
//                      user1.gno !> r/crossrealm_test.main2 !> r/demo/tests.ExecSwitch -> std.PreviousRealm = gno.land/r/crossrealm_test
//                            user1.gno !> r/crossrealm_test.main2 -> p/demo/tests.Exec -> std.PreviousRealm = user1.gno
//                                user1.gno !> r/crossrealm_test.main2 -> r/crossrealm_test.getPreviousRealm = user1.gno
//      user1.gno !> r/crossrealm_test.main2 -> r/crossrealm_test.Exec -> r/crossrealm_test.getPreviousRealm = user1.gno
//           user1.gno !> r/crossrealm_test.main2 -> r/demo/tests.Exec -> r/crossrealm_test.getPreviousRealm = user1.gno
//     user1.gno !> r/crossrealm_test.main2 !> r/demo/tests.ExecSwitch -> r/crossrealm_test.getPreviousRealm = gno.land/r/crossrealm_test
//           user1.gno !> r/crossrealm_test.main2 -> p/demo/tests.Exec -> r/crossrealm_test.getPreviousRealm = user1.gno
//                                                      user1.gno !> r/crossrealm_test.main2 -> p/demo/tests = user1.gno
//                            user1.gno !> r/crossrealm_test.main2 -> r/crossrealm_test.Exec -> p/demo/tests = user1.gno
//                                 user1.gno !> r/crossrealm_test.main2 -> r/demo/tests.Exec -> p/demo/tests = user1.gno
//                           user1.gno !> r/crossrealm_test.main2 !> r/demo/tests.ExecSwitch -> p/demo/tests = gno.land/r/crossrealm_test
//                                 user1.gno !> r/crossrealm_test.main2 -> p/demo/tests.Exec -> p/demo/tests = user1.gno
//                             user1.gno !> r/crossrealm_test.main2 -> p/demo/tests -> p/demo/tests/subtests = user1.gno
//   user1.gno !> r/crossrealm_test.main2 -> r/crossrealm_test.Exec -> p/demo/tests -> p/demo/tests/subtests = user1.gno
//        user1.gno !> r/crossrealm_test.main2 -> r/demo/tests.Exec -> p/demo/tests -> p/demo/tests/subtests = user1.gno
//  user1.gno !> r/crossrealm_test.main2 !> r/demo/tests.ExecSwitch -> p/demo/tests -> p/demo/tests/subtests = gno.land/r/crossrealm_test
//        user1.gno !> r/crossrealm_test.main2 -> p/demo/tests.Exec -> p/demo/tests -> p/demo/tests/subtests = user1.gno
//                                                      user1.gno !> r/crossrealm_test.main2 !> r/demo/tests = gno.land/r/crossrealm_test
//                            user1.gno !> r/crossrealm_test.main2 -> r/crossrealm_test.Exec !> r/demo/tests = gno.land/r/crossrealm_test
//                                 user1.gno !> r/crossrealm_test.main2 -> r/demo/tests.Exec !> r/demo/tests = gno.land/r/crossrealm_test
//                           user1.gno !> r/crossrealm_test.main2 !> r/demo/tests.ExecSwitch !> r/demo/tests = gno.land/r/demo/tests
//                                 user1.gno !> r/crossrealm_test.main2 -> p/demo/tests.Exec !> r/demo/tests = gno.land/r/crossrealm_test
//                             user1.gno !> r/crossrealm_test.main2 !> r/demo/tests !> r/demo/tests/subtests = gno.land/r/demo/tests
//   user1.gno !> r/crossrealm_test.main2 -> r/crossrealm_test.Exec !> r/demo/tests !> r/demo/tests/subtests = gno.land/r/demo/tests
//        user1.gno !> r/crossrealm_test.main2 -> r/demo/tests.Exec !> r/demo/tests !> r/demo/tests/subtests = gno.land/r/demo/tests
//  user1.gno !> r/crossrealm_test.main2 !> r/demo/tests.ExecSwitch !> r/demo/tests !> r/demo/tests/subtests = gno.land/r/demo/tests
//        user1.gno !> r/crossrealm_test.main2 -> p/demo/tests.Exec !> r/demo/tests !> r/demo/tests/subtests = gno.land/r/demo/tests
