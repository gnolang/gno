package crossrealm_c

import (
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
	println("Owner from crossrealm EntryPoint noCrossing: " + owner)
	println()
	owner2 := crossrealm.EntryPointWithCrossing().Owner()
	println("Owner from crossrealm EntryPoint crossing: " + owner2)
	println()
	owner = crossrealm.PrevRealmNoCrossing().Owner()
	println("Owner from crossrealm PrevRealmNoCrossing: " + owner)
	println()
	owner = cross(crossrealm.PrevRealmCrossing)().Owner()
	println("Owner from crossrealm PrevRealmCrossing: " + owner)
}
