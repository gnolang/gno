package crossrealm_c

import (
	"chain/runtime"

	"gno.land/p/demo/tests/p_crossrealm"
	"gno.land/r/tests/vm/crossrealm"
)

// EntryPoint is a non-crossing helper. Callers thread their own
// live cur as `rlm`; `passThrough` and the downstream `cross2(rlm)`
// calls forward it into crossrealm's crossing functions.
func EntryPoint(_ int, rlm realm) {
	println("crossrealm_c  EntryPoint: " + runtime.PreviousRealm().PkgPath())
	println("crossrealm_c  EntryPoint: " + runtime.PreviousRealm().Address())
	println(" ")
	passThrough(0, rlm)
}

func passThrough(_ int, rlm realm) {
	owner := crossrealm.EntryPoint().Owner()
	println("Owner from crossrealm EntryPoint no crossing: " + owner)
	println()
	owner2 := crossrealm.EntryPointWithCrossing(0, rlm).Owner()
	println("Owner from crossrealm EntryPoint with crossing: " + owner2)
	println()
	owner = crossrealm.PrevRealmNoCrossing().Owner()
	println("Owner from crossrealm PrevRealmNoCrossing: " + owner)
	println()
	owner = crossrealm.PrevRealmCrossing(cross2(rlm)).Owner()
	println("Owner from crossrealm PrevRealmCrossing: " + owner)
}

func CurRealmEntryPoint(_ int, rlm realm) {

	println("crossrealm CurRealmWithCrossing: " + crossrealm.CurRealmCrossing(cross2(rlm)).Address())
	println("crossrealm CurRealmNoCrossing: " + crossrealm.CurRealmNoCrossing().Address())

	println("p_crossrealm CurRealm: " + p_crossrealm.CurrentRealm().Address())
	println("crossrealm PkgCurRealmWithCrossing: " + crossrealm.PkgCurRealmCrossing(cross2(rlm)).Address())
	println("crossrealm PkgCurRealmNoCrossing: " + crossrealm.PkgCurRealmNoCrossing().Address())

}
