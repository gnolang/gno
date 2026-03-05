package registry

import (
	"errors"

	"gno.land/p/nt/ownable"
)

var (
	mainAddr   address
	backupAddr address
	owner      *ownable.Ownable
)

func init() {
	mainAddr = "g1sd5ezmxt4rwpy52u6wl3l3y085n8x0p6nllxm8"
	backupAddr = "g13awn2575t8s2vf3svlprc4dg0e9z5wchejdxk8"

	owner = ownable.NewWithAddressByPrevious(mainAddr)
}

func MainAddr() address {
	return mainAddr
}

func BackupAddr() address {
	return backupAddr
}

func SetMainAddr(_ realm, addr address) error {
	if !addr.IsValid() {
		return errors.New("config: invalid address")
	}

	owner.AssertOwned()

	mainAddr = addr
	return nil
}

func SetBackupAddr(_ realm, addr address) error {
	if !addr.IsValid() {
		return errors.New("config: invalid address")
	}

	owner.AssertOwned()

	backupAddr = addr
	return nil
}
