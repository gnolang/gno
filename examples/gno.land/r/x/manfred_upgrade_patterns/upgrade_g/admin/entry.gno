package admin 

import (
	"std"
)

var ReadCounter func() uint64 = nil
var UpdateCounter func() = nil
var DebugPrevRealm func() string = nil

// ----------------------------------------------------------------------------

func DebugAdminPrevRealm() string {
	return std.PrevRealm().Addr().String()
}
