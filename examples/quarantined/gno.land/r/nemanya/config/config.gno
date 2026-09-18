package config

import "errors"

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

func SetAddress(cur realm, a address) error {
	if !a.IsValid() {
		return ErrInvalidAddr
	}

	if err := checkAuthorized(0, cur); err != nil {
		return err
	}

	main = a
	return nil
}

func SetBackup(cur realm, a address) error {
	if !a.IsValid() {
		return ErrInvalidAddr
	}

	if err := checkAuthorized(0, cur); err != nil {
		return err
	}

	backup = a
	return nil
}

func checkAuthorized(_ int, rlm realm) error {
	if !rlm.IsCurrent() {
		return ErrUnauthorized
	}
	caller := rlm.Previous().Address()
	isAuthorized := caller == main || caller == backup

	if !isAuthorized {
		return ErrUnauthorized
	}

	return nil
}
