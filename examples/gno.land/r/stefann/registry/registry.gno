package registry

import (
	"chain"
	"errors"

	"gno.land/p/demo/ownable"
)

var (
	mainAddr   chain.Address
	backupAddr chain.Address
	owner      *ownable.Ownable
)

func init() {
	mainAddr = "g1sd5ezmxt4rwpy52u6wl3l3y085n8x0p6nllxm8"
	backupAddr = "g13awn2575t8s2vf3svlprc4dg0e9z5wchejdxk8"

	owner = ownable.NewWithAddress(mainAddr)
}

func MainAddr() chain.Address {
	return mainAddr
}

func BackupAddr() chain.Address {
	return backupAddr
}

func SetMainAddr(addr chain.Address) error {
	if !addr.IsValid() {
		return errors.New("config: invalid address")
	}

	owner.AssertCallerIsOwner()

	mainAddr = addr
	return nil
}

func SetBackupAddr(addr chain.Address) error {
	if !addr.IsValid() {
		return errors.New("config: invalid address")
	}

	owner.AssertCallerIsOwner()

	backupAddr = addr
	return nil
}
