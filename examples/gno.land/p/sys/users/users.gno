package users

import (
	"std"

	"gno.land/p/demo/ufmt"
)

type User struct {
	Address std.Address
	Name    string
}

func NewUser(address std.Address, name string) *User {
	return &User{
		Address: address,
		Name:    name,
	}
}

func (u *User) Render() string {
	out := ufmt.Sprintf("# User %s\n\n", u.Name)
	out += ufmt.Sprintf("- Address: %s\n\n", u.Address.String())

	return out
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
