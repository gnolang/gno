package boards2

import (
	"std"

	"gno.land/p/demo/boards2/admindao"
	"gno.land/r/demo/users"
)

func handleBoardCreate(args Args, _ *admindao.AdminDAO, cb func(Args)) {
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
