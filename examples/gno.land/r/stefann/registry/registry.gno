package registry

import (
	"errors"
	"std"

	"gno.land/p/demo/ownable"
)

var (
	mainAddr   std.Address
	backupAddr std.Address
	owner      *ownable.Ownable
)

func init() {
	mainAddr = "g1sd5ezmxt4rwpy52u6wl3l3y085n8x0p6nllxm8"
	backupAddr = "g13awn2575t8s2vf3svlprc4dg0e9z5wchejdxk8"

	owner = ownable.NewWithAddress(mainAddr)
}

func MainAddr() std.Address {
	return mainAddr
}

func BackupAddr() std.Address {
	return backupAddr
}

func SetMainAddr(_ realm, addr std.Address) error {
	if !addr.IsValid() {
		return errors.New("config: invalid address")
	}

	owner.AssertOwnedByPrevious()

	mainAddr = addr
	return nil
}

func SetBackupAddr(_ realm, addr std.Address) error {
	if !addr.IsValid() {
		return errors.New("config: invalid address")
	}

	owner.AssertOwnedByPrevious()

	backupAddr = addr
	return nil
}
