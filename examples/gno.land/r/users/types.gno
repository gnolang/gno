package users

type AddressOrName string

func (aon AddressOrName) IsName() bool {
	return aon[0] == '@'
}

func (aon AddressOrName) GetName() (string, bool) {
	if aon[0] == '@' {
		return aon[1:], true
	} else {
		return "", false
	}
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
