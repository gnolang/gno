package registry

import (
	"errors"
	"std"
)

var (
	mainAddress   std.Address
	backupAddress std.Address

	ErrInvalidAddr  = errors.New("Ivan's registry: Invalid address")
	ErrUnauthorized = errors.New("Ivan's registry: Unauthorized")
)

func init() {
	mainAddress = "g1d24j8fwnc0w5q427fauyey4gdd30qgu69k6n0x"
	backupAddress = "g1mw2xft3eava9kfhqw3fjj3kkf3pkammty0mtv7"
}

func MainAddress() std.Address {
	return mainAddress
}

func BackupAddress() std.Address {
	return backupAddress
}

func SetMainAddress(_ realm, addr std.Address) error {
	assertAuthorized()

	if !addr.IsValid() {
		return ErrInvalidAddr
	}

	mainAddress = addr
	return nil
}

func SetBackupAddress(_ realm, addr std.Address) error {
	assertAuthorized()

	if !addr.IsValid() {
		return ErrInvalidAddr
	}

	backupAddress = addr
	return nil
}

// It will stay here for now, might be useful later
func assertAuthorized() {
	caller := std.PreviousRealm().Address()
	isAuthorized := caller == mainAddress || caller == backupAddress

	if !isAuthorized {
		panic(ErrUnauthorized)
	}
}
