package crossrealm_c

import (
	"gno.land/p/demo/tests/p_crossrealm"
	"gno.land/r/demo/tests/crossrealm"
	"std"
)

func EntryPoint() {
	println("crossrealm_c  EntryPoint: " + std.PreviousRealm().PkgPath())
	println("crossrealm_c  EntryPoint: " + std.PreviousRealm().Address())
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
	owner = cross(crossrealm.PrevRealmCrossing)().Owner()
	println("Owner from crossrealm PrevRealmCrossing: " + owner)
}

func CurRealmEntryPoint() {

	println("crossrealm CurRealmWithCrossing: " + cross(crossrealm.CurRealmWithCrossing)().Address())
	println("crossrealm CurRealmNoCrossing: " + crossrealm.CurRealmNoCrossing().Address())

	println("p_crossrealm CurRealm: " + p_crossrealm.CurrentRealm().Address())
	println("crossrealm PkgCurRealmWithCrossing: " + cross(crossrealm.PkgCurRealmWithCrossing)().Address())
	println("crossrealm PkgCurRealmNoCrossing: " + crossrealm.PkgCurRealmNoCrossing().Address())

}
