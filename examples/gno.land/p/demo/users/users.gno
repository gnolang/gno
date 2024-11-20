package users

import (
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
)

//----------------------------------------
// Types

type User struct {
	Address std.Address
	Name    string
	Profile string
	Number  int
	Block   int
}

func (u *User) Render() string {
	out := ufmt.Sprintf("# User - %s\n\n", u.Name)

	out += "## Bio\n\n"
	if strings.TrimSpace(u.Profile) != "" {
		out += ufmt.Sprintf("%s\n\n", u.Profile)
	} else {
		out += "No bio provided.\n\n"
	}

	out += "## Stats\n\n"
	out += ufmt.Sprintf("- ID #%d\n\n", u.Number)
	out += ufmt.Sprintf("- Address: %s\n\n", u.Address.String())
	out += ufmt.Sprintf("- Registered at block #%d\n\n", u.Block)

	return out
}
