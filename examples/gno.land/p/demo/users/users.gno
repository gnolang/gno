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
	Profile string // todo rename to bio
	Number  int    // todo rename to ID, switch to uint or seqid? XXX: do we really need this?
	Block   int64  // XXX: should we save this?
}

func (u *User) Render() string {
	out := ufmt.Sprintf("# User #%d - %s\n\n", u.Number, u.Name)

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
