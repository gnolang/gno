package home

import (
	"errors"
	"std"
	"gno.land/p/moul/authz"
)

var (
	mainAddr   = std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y") // matija's main address
	backupAddr std.Address                                               // backup address

	errorInvalidAddr  = errors.New("config: invalid address")
	errorUnauthorized = errors.New("config: unauthorized")

	Auth *authz.Authorizer
)

func init() {
	Auth = authz.NewWithAuthority(authz.NewMemberAuthority(mainAddr, backupAddr))
}

func Address() std.Address {
	return mainAddr
}

func Backup() std.Address {
	return backupAddr
}

func SetAddress(newAddress std.Address) error {
	if !newAddress.IsValid() {
		return errorInvalidAddr
	}

	if err := Auth.DoByPrevious("set_address", func() error {
		mainAddr = newAddress
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func SetBackup(newAddress std.Address) error {
	if !newAddress.IsValid() {
		return errorInvalidAddr
	}

	if err := Auth.DoByPrevious("set_backup", func() error {
		backupAddr = newAddress
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func GetAuth() *authz.Authorizer {
	return Auth
}
