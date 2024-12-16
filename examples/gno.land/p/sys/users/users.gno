package users

import (
	"std"
)

type UserLookup struct {
	Name  string
	Addr  std.Address
	Alias bool // lookup was found via an alias
}

func NewUserLookup(name string, addr std.Address, alias bool) UserLookup {
	return UserLookup{
		Name:  name,
		Addr:  addr,
		Alias: alias,
	}
}

func (lookup UserLookup) Exists() bool {
	return lookup.Addr != ""
}

func (lookup UserLookup) Deleted() bool {
	return lookup.Addr == "" && lookup.Name == "ghost" // todo see "ghost"?
}

// AddressOrName can be used for resolving either the address or the username
type AddressOrName string

func (aon AddressOrName) IsName() bool {
	return aon != "" && aon[0] == '@'
}

func (aon AddressOrName) Name() (string, bool) {
	if len(aon) >= 2 && aon[0] == '@' {
		return string(aon[1:]), true
	}

	return "", false
}
