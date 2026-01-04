package registry

import (
	"chain/runtime"
	"errors"
)

var (
	mainAddress   address
	backupAddress address

	ErrInvalidAddr  = errors.New("Ivan's registry: Invalid address")
	ErrUnauthorized = errors.New("Ivan's registry: Unauthorized")
)

func init() {
	mainAddress = "g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x"
	backupAddress = "g1mw2xft3eava9kfhqw3fjj3kkf3pkammty0mtv7"
}

func MainAddress() address {
	return mainAddress
}

func BackupAddress() address {
	return backupAddress
}

func SetMainAddress(_ realm, addr address) error {
	assertAuthorized()

	if !addr.IsValid() {
		return ErrInvalidAddr
	}

	mainAddress = addr
	return nil
}

func SetBackupAddress(_ realm, addr address) error {
	assertAuthorized()

	if !addr.IsValid() {
		return ErrInvalidAddr
	}

	backupAddress = addr
	return nil
}

// It will stay here for now, might be useful later
func assertAuthorized() {
	caller := runtime.PreviousRealm().Address()
	isAuthorized := caller == mainAddress || caller == backupAddress

	if !isAuthorized {
		panic(ErrUnauthorized)
	}
}
