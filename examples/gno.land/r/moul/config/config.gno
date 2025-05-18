package config

import (
	"gno.land/p/moul/authz"
)

var (
	Authority, privConfig = authz.New("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // XXX: when #4264 -> authz.NewWithOrigin()
	SafeConfig            = privConfig.Safe()
)

func PrivilegedAction() {
	crossing()
	Authority.AssertPreviousCan("call_privileged_action")
	std.Emit("PrivilegedAction called",
		"by", std.PreviousRealm().Address().String(),
	)
}
