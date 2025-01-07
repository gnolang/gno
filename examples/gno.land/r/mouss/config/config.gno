package config

import (
	"errors"
	"std"
)

var (
	main   std.Address // mouss's main address
	backup std.Address // backup address

	errorInvalidAddr  = errors.New("config: invalid address")
	errorUnauthorized = errors.New("config: unauthorized")
)

func init() {
	main = "g1hrfvdh7jdvnlxpk2y20tp3scj9jqal3zzu7wjz"
}
func Address() std.Address {
	return main
}

func Backup() std.Address {
	return backup
}
func SetBack(buAddress std.Address) error {
	if !buAddress.IsValid() {
		return errorInvalidAddr
	}

	if err := checkAuthorized(); err != nil {
		return err
	}

	backup = buAddress
	return nil
}

func checkAuthorized() error {
	caller := std.GetOrigCaller()
	if caller != main && caller != backup {
		return errorUnauthorized
	}

	return nil
}
