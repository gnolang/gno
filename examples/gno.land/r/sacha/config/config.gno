package config

import (
	"errors"

	"gno.land/p/nt/ownable"
)

var (
	OwnableMain   = ownable.NewWithAddressByPrevious("g1jpg2hmsh8j4f60u4vkseqh5azeupns8vhrstfm")
	OwnableBackup = ownable.NewWithAddressByPrevious("g1ez7rzp7dms7he05yy9nwdczl97qw233z892hsm")

	ErrUnauthorized = errors.New("You're not Sacha")
)

func IsAuthorized(addr address) bool {
	return addr == OwnableMain.Owner() || addr == OwnableBackup.Owner()
}
