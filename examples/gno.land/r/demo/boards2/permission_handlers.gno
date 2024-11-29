package boards2

import (
	"std"

	"gno.land/r/demo/users"
)

func handleBoardCreate(_ Permissioner, args Args, cb func(Args)) {
	// TODO: This way of dealing with arguments is delicate, ideally types should be used
	name := args[0].(string)
	if std.Address(name).IsValid() {
		panic("addresses are not allowed as board name")
	}

	// When the board name is the name of a registered user
	// check that caller is the owner of the name.
	caller := std.GetOrigCaller()
	user := users.GetUserByName(name)
	if user != nil && user.Address != caller {
		panic("board name is a user name registered to a different user")
	}

	cb(args)
}

func handleMemberInvite(p Permissioner, args Args, cb func(Args)) {
	// Make sure that only owners invite other owners
	role := args[1].(Role)
	if role == RoleOwner {
		caller := std.GetOrigCaller()
		if !p.HasRole(caller, RoleOwner) {
			panic("only owners are allowed to invite other owners")
		}
	}

	cb(args)
}
