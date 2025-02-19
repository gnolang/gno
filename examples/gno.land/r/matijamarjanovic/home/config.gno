package home

import (
	"errors"
	"std"
)

var (
	mainAddr   = std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y") // matija's main address
	backupAddr std.Address                                               // backup address

	errorInvalidAddr  = errors.New("config: invalid address")
	errorUnauthorized = errors.New("config: unauthorized")
)

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

	if err := checkAuthorized(); err != nil {
		return err
	}

	mainAddr = newAddress
	return nil
}

func SetBackup(newAddress std.Address) error {
	if !newAddress.IsValid() {
		return errorInvalidAddr
	}

	if err := checkAuthorized(); err != nil {
		return err
	}

	backupAddr = newAddress
	return nil
}

func checkAuthorized() error {
	caller := std.OriginCaller()
	if caller != mainAddr && caller != backupAddr {
		return errorUnauthorized
	}

	return nil
}

func AssertAuthorized() {
	caller := std.OriginCaller()
	if caller != mainAddr && caller != backupAddr {
		panic(errorUnauthorized)
	}
}
