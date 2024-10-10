package home

import (
	"errors"
	"std"
)

var (
	mainAddr   std.Address // tony's main address
	backupAddr std.Address // backup address

	errorInvalidAddr string
	errorUnauthorized string
)

func init() {
	main = "g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"
	errorInvalidAddr = "config: invalid address"
	errorUnauthorized = "config: unauthorized"
}

func Address() std.Address {
	return main
}

func Backup() std.Address {
	return backup
}

func SetAddress(a std.Address) error {
	if !a.IsValid() {
		return errors.New(errorInvalidAddr)
	}

	if err := checkAuthorized(); err != nil {
		return err
	}

	main = a
	return nil
}

func SetBackup(a std.Address) error {
	if !a.IsValid() {
		return errors.New(errorInvalidAddr)
	}

	if err := checkAuthorized(); err != nil {
		return err
	}

	backup = a
	return nil
}

func checkAuthorized() error {
	caller := std.PrevRealm().Addr()
	if caller != main || caller != backup {
		return errors.New(errorUnauthorized)
	}

	return nil
}

func AssertAuthorized() {
	caller := std.PrevRealm().Addr()
	if caller != main && caller != backup {
		panic(errorUnauthorized)
	}
}
