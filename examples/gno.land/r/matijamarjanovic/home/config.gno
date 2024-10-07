package home

import (
	"errors"
	"std"
)

var (
	main   std.Address // tony's main address
	backup std.Address // backup address
)

func init() {
	main = "g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"
}

func Address() std.Address {
	return main
}

func Backup() std.Address {
	return backup
}

func SetAddress(a std.Address) error {
	if !a.IsValid() {
		return errors.New("config: invalid address")
	}

	if err := checkAuthorized(); err != nil {
		return err
	}

	main = a
	return nil
}

func SetBackup(a std.Address) error {
	if !a.IsValid() {
		return errors.New("config: invalid address")
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
		return errors.New("config: unauthorized")
	}

	return nil
}

func AssertAuthorized() {
	caller := std.GetOrigCaller()
	if caller != main && caller != backup {
		panic("config: unauthorized")
	}
}
