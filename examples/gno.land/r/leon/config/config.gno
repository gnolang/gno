package config

import (
	"errors"
	"std"
)

var (
	main   std.Address // leon's main address
	backup std.Address // backup address

	ErrInvalidAddr  = errors.New("leon's config: invalid address")
	ErrUnauthorized = errors.New("leon's config: unauthorized")
)

func init() {
	main = "g125em6arxsnj49vx35f0n0z34putv5ty3376fg5"
}

func Address() std.Address {
	return main
}

func Backup() std.Address {
	return backup
}

func SetAddress(a std.Address) error {
	if !a.IsValid() {
		return ErrInvalidAddr
	}

	if err := checkAuthorized(); err != nil {
		return err
	}

	main = a
	return nil
}

func SetBackup(a std.Address) error {
	if !a.IsValid() {
		return ErrInvalidAddr
	}

	if err := checkAuthorized(); err != nil {
		return err
	}

	backup = a
	return nil
}

func checkAuthorized() error {
	caller := std.PrevRealm().Addr()
	isAuthorized := caller == main || caller == backup

	if !isAuthorized {
		return ErrUnauthorized
	}

	return nil
}
