package home

import (
	"chain"
	"chain/runtime"

	"gno.land/p/moul/authz"
)

var (
	mainAddr   = chain.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")
	backupAddr = chain.Address("g1kjsl2ungmc95mgluq96w8dqlep8d4n8cxdfthk")
	auth       = authz.NewMemberAuthority(mainAddr, backupAddr)
	Auth       = authz.NewWithAuthority(auth)
)

func Authorize(cur realm, addr string) {
	auth.AddMember(runtime.PreviousRealm().Address(), chain.Address(addr))
}

func Address() chain.Address {
	return mainAddr
}

func Backup() chain.Address {
	return backupAddr
}
