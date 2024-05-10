package diff

import (
	"std"
)

func PrintInfo() {
	println("diff info")
	println("currentRealm:", std.CurrentRealm().PkgPath())
}
