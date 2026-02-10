package config

import (
	"errors"

	"gno.land/p/demo/ownable"
)

var (
	OwnableMain   = ownable.NewWithAddress("g1wq2h93ppkf2gkgncz5unayrsmt7pl8npktnznd")
	OwnableBackup = ownable.NewWithAddress("g1hrfvdh7jdvnlxpk2y20tp3scj9jqal3zzu7wjz")

	ErrUnauthorized = errors.New("config: unauthorized")
)

func SetMainAddr(addr address) error {
	return OwnableMain.TransferOwnership(addr)
}

func SetBackupAddr(addr address) error {
	return OwnableBackup.TransferOwnership(addr)
}

func IsAuthorized(addr address) bool {
	return addr == OwnableMain.Owner() || addr == OwnableBackup.Owner()
}

func Render(path string) string {
	out := "# mouss configuration\n\n"

	out += "## Authorized Addresses\n\n"
	out += "- main: " + OwnableMain.Owner().String() + "\n"
	out += "- backup: " + OwnableBackup.Owner().String() + "\n\n"

	return out
}
