package config

import (
	"chain"

	"gno.land/p/moul/authz"
)

var (
	GnoDevRels = authz.NewMemberAuthority(
		chain.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5"), // @moul
		chain.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5"), // @leohhhn
	)
	Auth = authz.NewWithAuthority(GnoDevRels)
)

func Render(_ string) string {
	return Auth.Authority().String()
}
