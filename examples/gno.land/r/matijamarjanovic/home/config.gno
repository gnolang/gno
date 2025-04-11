package home

import (
	"chain"
	"chain/runtime"
	"errors"
)

var (
	mainAddr   = chain.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y") // matija's main address
	backupAddr chain.Address                                               // backup address

	errorInvalidAddr  = errors.New("config: invalid address")
	errorUnauthorized = errors.New("config: unauthorized")
)

func Address() chain.Address {
	return mainAddr
}

func Backup() chain.Address {
	return backupAddr
}

func SetAddress(newAddress chain.Address) error {
	if !newAddress.IsValid() {
		return errorInvalidAddr
	}

	if err := checkAuthorized(); err != nil {
		return err
	}

	mainAddr = newAddress
	return nil
}

func SetBackup(newAddress chain.Address) error {
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
	caller := runtime.OriginCaller()
	if caller != mainAddr && caller != backupAddr {
		return errorUnauthorized
	}

	return nil
}

func AssertAuthorized() {
	caller := runtime.OriginCaller()
	if caller != mainAddr && caller != backupAddr {
		panic(errorUnauthorized)
	}
}
