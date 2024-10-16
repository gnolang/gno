package teams

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/users"
)

type Team struct {
	Name	string
	Owners	[]std.Address
	Members	[]TeamMember
	Permissions	[]string
}

type TeamMember	struct {
	User	users.User,
	Permissions	[]string
}