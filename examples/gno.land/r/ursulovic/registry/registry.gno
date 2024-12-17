package registry

import (
	"errors"
	"std"
)

var (
	mainAddress   std.Address
	backupAddress std.Address

	ErrInvalidAddr  = errors.New("leon's config: invalid address")
	ErrUnauthorized = errors.New("leon's config: unauthorized")
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

func SetMainAddress(addr std.Address) error {
	if !addr.IsValid() {
		return ErrInvalidAddr
	}

	if err := checkAuthorized; err != nil {
		return err
	}
	mainAddress = addr
	return nil
}

func SetBackupAddress(addr std.Address) error {
	if !addr.IsValid() {
		return ErrInvalidAddr
	}

	if err := checkAuthorized; err != nil {
		return err
	}
	backupAddr = addr
	return nil
}

func assertAuthorized() {
	caller := std.PrevRealm().Addr()
	isAuthorized := caller == mainAddress || caller == backupAddress

	if !isAuthorized {
		panic(ErrUnauthorized)
	}
}

func checkAuthorized() error {
	caller := std.PrevRealm().Addr()
	isAuthorized := caller == mainAddress || caller == backupAddress

	if !isAuthorized {
		return ErrUnauthorized
	}
	return nil
}