// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"std"

	ptests "gno.land/p/demo/tests"
	"gno.land/p/nt/ufmt"
	rtests "gno.land/r/tests/vm"
)

func getPreviousRealm() std.Realm {
	return std.PreviousRealm()
}

func Exec(fn func()) {
	fn()
}

func main(cur realm) {
	// Create a map of the potential callers, this will give more understandable
	// output than the bech32 addresses.
	callersByAddr := make(map[std.Address]string)
	for _, caller := range []string{
		"user1.gno", "gno.land/r/crossrealm", tests "gno.land/r/tests/vm",
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
		callStackAdd     string
		callerFn         func() std.Realm
		callerFnCrossing func(realm) std.Realm
	}{
		{
			callStackAdd: " -> std.PreviousRealm",
			callerFn:     std.PreviousRealm,
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
			var r std.Realm
			if tt.callerFnCrossing != nil {
				r = tt.callerFnCrossing(cross)
			} else {
				r = tt.callerFn()
			}
			printColumns(baseCallStack+tt.callStackAdd, callersByAddr[r.Address()])

		}
		Exec(func() {
			var r std.Realm
			if tt.callerFnCrossing != nil {
				r = tt.callerFnCrossing(cross)
			} else {
				r = tt.callerFn()
			}
			assertRealm(r)
			printColumns(baseCallStack+" -> r/crossrealm.Exec"+tt.callStackAdd, callersByAddr[r.Address()])
		})
		rtests.Exec(func() {
			var r std.Realm
			if tt.callerFnCrossing != nil {
				r = tt.callerFnCrossing(cross)
			} else {
				r = tt.callerFn()
			}
			assertRealm(r)
			printColumns(baseCallStack+" -> r/demo/tests.Exec"+tt.callStackAdd, callersByAddr[r.Address()])
		})
		rtests.ExecSwitch(cross, func() {
			var r std.Realm
			if tt.callerFnCrossing != nil {
				r = tt.callerFnCrossing(cross)
			} else {
				r = tt.callerFn()
			}
			assertRealm(r)
			printColumns(baseCallStack+" !> r/demo/tests.ExecSwitch"+tt.callStackAdd, callersByAddr[r.Address()])
		})
		ptests.Exec(func() {
			var r std.Realm
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

// Error:
// zrealm_crossrealm11.gno:25:47: missing ',' in composite literal (and 3 more errors)

// TypeCheckError:
// gno.land/r/crossrealm/zrealm_crossrealm11.gno:25:47: missing ',' in composite literal (and 3 more errors)
