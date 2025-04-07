package config

import (
	"std"

	"gno.land/p/moul/authz"
)

var GnoDevRels = authz.NewMemberAuthority(
	std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5"), // @moul
	std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5"), // @leohhhn
)

func Render(_ string) string {
	return GnoDevRels.String()
}
