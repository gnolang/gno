package config

import (
	"chain/runtime"
	"errors"
)

var (
	main   address
	backup address

	ErrInvalidAddr  = errors.New("Invalid address")
	ErrUnauthorized = errors.New("Unauthorized")
)

func init() {
	main = "g1x9qyf6f34v2g52k4q5smn5tctmj3hl2kj7l2ql"
}

func Address() address {
	return main
}

func Backup() address {
	return backup
}

func SetAddress(_ realm, a address) error {
	if !a.IsValid() {
		return ErrInvalidAddr
	}

	if err := checkAuthorized(); err != nil {
		return err
	}

	main = a
	return nil
}

func SetBackup(_ realm, a address) error {
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
	caller := runtime.PreviousRealm().Address()
	isAuthorized := caller == main || caller == backup

	if !isAuthorized {
		return ErrUnauthorized
	}

	return nil
}
