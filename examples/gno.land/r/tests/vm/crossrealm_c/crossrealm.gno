package crossrealm_c

import (
	"chain/runtime"

	"gno.land/p/demo/tests/p_crossrealm"
	"gno.land/r/tests/vm/crossrealm"
)

func EntryPoint() {
	println("crossrealm_c  EntryPoint: " + runtime.PreviousRealm().PkgPath())
	println("crossrealm_c  EntryPoint: " + runtime.PreviousRealm().Address())
	println(" ")
	passThrough()

}

func passThrough() {
	owner := crossrealm.EntryPoint().Owner()
	println("Owner from crossrealm EntryPoint no crossing: " + owner)
	println()
	owner2 := crossrealm.EntryPointWithCrossing().Owner()
	println("Owner from crossrealm EntryPoint with crossing: " + owner2)
	println()
	owner = crossrealm.PrevRealmNoCrossing().Owner()
	println("Owner from crossrealm PrevRealmNoCrossing: " + owner)
	println()
	owner = crossrealm.PrevRealmCrossing(cross).Owner()
	println("Owner from crossrealm PrevRealmCrossing: " + owner)
}

func CurRealmEntryPoint() {

	println("crossrealm CurRealmWithCrossing: " + crossrealm.CurRealmCrossing(cross).Address())
	println("crossrealm CurRealmNoCrossing: " + crossrealm.CurRealmNoCrossing().Address())

	println("p_crossrealm CurRealm: " + p_crossrealm.CurrentRealm().Address())
	println("crossrealm PkgCurRealmWithCrossing: " + crossrealm.PkgCurRealmCrossing(cross).Address())
	println("crossrealm PkgCurRealmNoCrossing: " + crossrealm.PkgCurRealmNoCrossing().Address())

}
