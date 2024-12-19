package users

import (
	"std"
)

type UserLookup struct {
	Name     string
	Addr     std.Address
	ViaAlias bool // lookup was found via an alias
}

func (lookup UserLookup) Exists() bool {
	return lookup.Addr != ""
}

func (lookup UserLookup) Deleted() bool {
	return lookup.Addr == "" && lookup.Name == "ghost" // todo see "ghost"?
}
