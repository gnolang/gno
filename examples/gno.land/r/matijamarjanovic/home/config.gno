package home

import (
	"std"

	"gno.land/p/moul/authz"
)

var (
	mainAddr   = std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")
	backupAddr = std.Address("g1kjsl2ungmc95mgluq96w8dqlep8d4n8cxdfthk")
	auth       = authz.NewMemberAuthority(mainAddr, backupAddr)
	Auth       = authz.NewWithAuthority(auth)
)

func Authorize(cur realm, addr string) {
	auth.AddMember(std.PreviousRealm().Address(), std.Address(addr))
}

func Address() std.Address {
	return mainAddr
}

func Backup() std.Address {
	return backupAddr
}
