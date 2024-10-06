package config

import (
	"errors"
	"std"
)

var (
	main   std.Address
	backup std.Address
)

func init() {
	main = "g1sd5ezmxt4rwpy52u6wl3l3y085n8x0p6nllxm8"
	backup = "g13awn2575t8s2vf3svlprc4dg0e9z5wchejdxk8"
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

func checkAuthorized() error {
	caller := std.PrevRealm().Addr()
	if caller != main || caller != backup {
		return errors.New("config: unauthorized")
	}

	return nil
}

func AssertAuthorized() {
	caller := std.PrevRealm().Addr()
	if caller != main || caller != backup {
		panic("config: unauthorized")
	}
}