package home

import (
	"chain/runtime"

	"gno.land/p/moul/authz"
)

var (
	mainAddr   = address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")
	backupAddr = address("g1kjsl2ungmc95mgluq96w8dqlep8d4n8cxdfthk")
	auth       = authz.NewMemberAuthority(mainAddr, backupAddr)
	Auth       = authz.NewWithAuthority(auth)
)

func Authorize(cur realm, addr string) {
	auth.AddMember(runtime.PreviousRealm().Address(), address(addr))
}

func Address() address {
	return mainAddr
}

func Backup() address {
	return backupAddr
}
