package config

import (
	"errors"
	"std"

	"gno.land/p/demo/ownable"
)

var (
	OwnableMain   = ownable.NewWithAddress("g1jpg2hmsh8j4f60u4vkseqh5azeupns8vhrstfm")
	OwnableBackup = ownable.NewWithAddress("g1ez7rzp7dms7he05yy9nwdczl97qw233z892hsm")

	ErrUnauthorized = errors.New("You're not Sacha")
)

func IsAuthorized(addr std.Address) bool {
	return addr == OwnableMain.Owner() || addr == OwnableBackup.Owner()
}
