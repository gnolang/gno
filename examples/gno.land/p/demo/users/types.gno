package users

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
