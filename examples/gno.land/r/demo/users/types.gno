package users

import "std"

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

func (aon AddressOrName) Resolve() std.Address {
	name, isName := aon.GetName()
	if isName {
		user := GetUserByName(name)
		return user.address
	} else {
		return std.Address(aon) // TODO check validity
	}
}
