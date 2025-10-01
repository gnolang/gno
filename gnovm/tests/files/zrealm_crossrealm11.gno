// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"chain"
	"chain/runtime"

	ptests "gno.land/p/demo/tests"
	"gno.land/p/nt/ufmt"
	rtests "gno.land/r/tests/vm"
)

func getPreviousRealm() runtime.Realm {
	return runtime.PreviousRealm()
}

func Exec(fn func()) {
	fn()
}

func main(cur realm) {
	// Create a map of the potential callers, this will give more understandable
	// output than the bech32 addresses.
	callersByAddr := make(map[address]string)
	for _, caller := range []string{
		"user1.gno", "gno.land/r/crossrealm", "gno.land/r/tests/vm",
	} {
		addr := chain.PackageAddress(caller)
		callersByAddr[addr] = caller
	}

	assertRealm := func(r runtime.Realm) {
		pkgPath := callersByAddr[r.Address()]
		if r.IsUser() && pkgPath != "user1.gno" {
			panic(ufmt.Sprintf("ERROR: expected: 'user1.gno', got:'%s'", pkgPath))
		} else if !r.IsUser() && pkgPath != r.PkgPath() {
			panic(ufmt.Sprintf("ERROR: expected: '%s', got: '%s'", pkgPath, r.PkgPath()))
		}
	}

	tests := []struct {
		callStackAdd     string
		callerFn         func() runtime.Realm
		callerFnCrossing func(realm) runtime.Realm
	}{
		{
			callStackAdd: " -> std.PreviousRealm",
			callerFn:     runtime.PreviousRealm,
		},
		{
			callStackAdd: " -> r/crossrealm.getPreviousRealm",
			callerFn:     getPreviousRealm,
		},
		{
			callStackAdd: " -> p/demo/tests",
			callerFn:     ptests.GetPreviousRealm,
		},
		{
			callStackAdd: " -> p/demo/tests -> p/demo/tests/subtests",
			callerFn:     ptests.GetPSubtestsPreviousRealm,
		},
		{
			callStackAdd:     " !> r/demo/tests",
			callerFnCrossing: rtests.GetPreviousRealm,
		},
		{
			callStackAdd:     " !> r/demo/tests !> r/demo/tests/subtests",
			callerFnCrossing: rtests.GetRSubtestsPreviousRealm,
		},
	}

	println("---") // needed to have space prefixes
	printColumns("STACK", "std.PreviousRealm")
	printColumns("-----", "------------------")

	baseCallStack := "user1.gno !> r/crossrealm.main2"
	for _, tt := range tests {
		{ // with no Exec
			var r runtime.Realm
			if tt.callerFnCrossing != nil {
				r = tt.callerFnCrossing(cross)
			} else {
				r = tt.callerFn()
			}
			printColumns(baseCallStack+tt.callStackAdd, callersByAddr[r.Address()])

		}
		Exec(func() {
			var r runtime.Realm
			if tt.callerFnCrossing != nil {
				r = tt.callerFnCrossing(cross)
			} else {
				r = tt.callerFn()
			}
			assertRealm(r)
			printColumns(baseCallStack+" -> r/crossrealm.Exec"+tt.callStackAdd, callersByAddr[r.Address()])
		})
		rtests.Exec(func() {
			var r runtime.Realm
			if tt.callerFnCrossing != nil {
				r = tt.callerFnCrossing(cross)
			} else {
				r = tt.callerFn()
			}
			assertRealm(r)
			printColumns(baseCallStack+" -> r/demo/tests.Exec"+tt.callStackAdd, callersByAddr[r.Address()])
		})
		rtests.ExecSwitch(cross, func() {
			var r runtime.Realm
			if tt.callerFnCrossing != nil {
				r = tt.callerFnCrossing(cross)
			} else {
				r = tt.callerFn()
			}
			assertRealm(r)
			printColumns(baseCallStack+" !> r/demo/tests.ExecSwitch"+tt.callStackAdd, callersByAddr[r.Address()])
		})
		ptests.Exec(func() {
			var r runtime.Realm
			if tt.callerFnCrossing != nil {
				r = tt.callerFnCrossing(cross)
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
//                                                      user1.gno !> r/crossrealm.main2 -> std.PreviousRealm = user1.gno
//                                 user1.gno !> r/crossrealm.main2 -> r/crossrealm.Exec -> std.PreviousRealm = user1.gno
//                                 user1.gno !> r/crossrealm.main2 -> r/demo/tests.Exec -> std.PreviousRealm = user1.gno
//                           user1.gno !> r/crossrealm.main2 !> r/demo/tests.ExecSwitch -> std.PreviousRealm = gno.land/r/crossrealm
//                                 user1.gno !> r/crossrealm.main2 -> p/demo/tests.Exec -> std.PreviousRealm = user1.gno
//                                          user1.gno !> r/crossrealm.main2 -> r/crossrealm.getPreviousRealm = user1.gno
//                     user1.gno !> r/crossrealm.main2 -> r/crossrealm.Exec -> r/crossrealm.getPreviousRealm = user1.gno
//                     user1.gno !> r/crossrealm.main2 -> r/demo/tests.Exec -> r/crossrealm.getPreviousRealm = user1.gno
//               user1.gno !> r/crossrealm.main2 !> r/demo/tests.ExecSwitch -> r/crossrealm.getPreviousRealm = gno.land/r/crossrealm
//                     user1.gno !> r/crossrealm.main2 -> p/demo/tests.Exec -> r/crossrealm.getPreviousRealm = user1.gno
//                                                           user1.gno !> r/crossrealm.main2 -> p/demo/tests = user1.gno
//                                      user1.gno !> r/crossrealm.main2 -> r/crossrealm.Exec -> p/demo/tests = user1.gno
//                                      user1.gno !> r/crossrealm.main2 -> r/demo/tests.Exec -> p/demo/tests = user1.gno
//                                user1.gno !> r/crossrealm.main2 !> r/demo/tests.ExecSwitch -> p/demo/tests = gno.land/r/crossrealm
//                                      user1.gno !> r/crossrealm.main2 -> p/demo/tests.Exec -> p/demo/tests = user1.gno
//                                  user1.gno !> r/crossrealm.main2 -> p/demo/tests -> p/demo/tests/subtests = user1.gno
//             user1.gno !> r/crossrealm.main2 -> r/crossrealm.Exec -> p/demo/tests -> p/demo/tests/subtests = user1.gno
//             user1.gno !> r/crossrealm.main2 -> r/demo/tests.Exec -> p/demo/tests -> p/demo/tests/subtests = user1.gno
//       user1.gno !> r/crossrealm.main2 !> r/demo/tests.ExecSwitch -> p/demo/tests -> p/demo/tests/subtests = gno.land/r/crossrealm
//             user1.gno !> r/crossrealm.main2 -> p/demo/tests.Exec -> p/demo/tests -> p/demo/tests/subtests = user1.gno
//                                                           user1.gno !> r/crossrealm.main2 !> r/demo/tests = gno.land/r/crossrealm
//                                      user1.gno !> r/crossrealm.main2 -> r/crossrealm.Exec !> r/demo/tests = gno.land/r/crossrealm
//                                      user1.gno !> r/crossrealm.main2 -> r/demo/tests.Exec !> r/demo/tests = gno.land/r/crossrealm
//                                user1.gno !> r/crossrealm.main2 !> r/demo/tests.ExecSwitch !> r/demo/tests = gno.land/r/tests/vm
//                                      user1.gno !> r/crossrealm.main2 -> p/demo/tests.Exec !> r/demo/tests = gno.land/r/crossrealm
//                                  user1.gno !> r/crossrealm.main2 !> r/demo/tests !> r/demo/tests/subtests = gno.land/r/tests/vm
//             user1.gno !> r/crossrealm.main2 -> r/crossrealm.Exec !> r/demo/tests !> r/demo/tests/subtests = gno.land/r/tests/vm
//             user1.gno !> r/crossrealm.main2 -> r/demo/tests.Exec !> r/demo/tests !> r/demo/tests/subtests = gno.land/r/tests/vm
//       user1.gno !> r/crossrealm.main2 !> r/demo/tests.ExecSwitch !> r/demo/tests !> r/demo/tests/subtests = gno.land/r/tests/vm
//             user1.gno !> r/crossrealm.main2 -> p/demo/tests.Exec !> r/demo/tests !> r/demo/tests/subtests = gno.land/r/tests/vm
