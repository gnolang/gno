package users

import (
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
)

type User struct {
	Address std.Address
	Name    string
	Bio     string
	Number  uint64
	Block   int64 // XXX: should we save this?
}

func (u *User) Render() string {
	out := ufmt.Sprintf("# User #%d - %s\n\n", u.Number, u.Name)

	out += "## Bio\n\n"
	if strings.TrimSpace(u.Bio) != "" {
		out += ufmt.Sprintf("%s\n\n", u.Bio)
	} else {
		out += "No bio provided.\n\n"
	}

	out += "## Stats\n\n"
	out += ufmt.Sprintf("- ID #%d\n\n", u.Number)
	out += ufmt.Sprintf("- Address: %s\n\n", u.Address.String())
	out += ufmt.Sprintf("- Registered at block #%d\n\n", u.Block)

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
