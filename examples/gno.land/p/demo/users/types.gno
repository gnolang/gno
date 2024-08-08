package users

import (
	"std"
)

type AddressOrName string

func (aon AddressOrName) IsName() bool {
	return aon != "" && aon[0] == '@'
}

func (aon AddressOrName) GetName() (string, bool) {
	if len(aon) >= 2 && aon[0] == '@' {
		return string(aon[1:]), true
	}
	return "", false
}

func (aon AddressOrName) IsValid() bool {
	if len(aon) == 0 {
		return false
	}
	if aon[0] == '@' {
		// TODO: validate name better
		return len(aon) >= 2
	}
	return std.Address(aon).IsValid()
}
