package users

import (
	"std"
	"strconv"
)

//----------------------------------------
// Types

type User struct {
	Address std.Address
	Name    string
	Profile string
	Number  int
	Invites int
	Inviter std.Address
}

func (u *User) Render() string {
	str := "## user " + u.Name + "\n" +
		"\n" +
		" * address = " + string(u.Address) + "\n" +
		" * " + strconv.Itoa(u.Invites) + " invites\n"
	if u.Inviter != "" {
		str = str + " * invited by " + string(u.Inviter) + "\n"
	}
	str = str + "\n" +
		u.Profile + "\n"
	return str
}
