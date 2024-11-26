package boards2

import (
	"errors"
	"std"

	"gno.land/p/demo/boards2/admindao"
	"gno.land/r/demo/users"
)

// Default board creation fee in ugnot required for anonymous users
const defaultAnonymousFee = 20 * 1_000_000 // TODO: Allow changing the fee though a proposal

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

	// Try to get the user by address or otheriwse consider the caller an anonymous user.
	// Anonymous users must pay a fee to create boards.
	if user == nil && users.GetUserByAddress(caller) == nil {
		sent := std.GetOrigSend()
		fee := std.NewCoin("ugnot", int64(defaultAnonymousFee))
		if len(sent) == 0 || sent[0].IsLT(fee) {
			panic("please register a user, otherwise a minimum fee of " + fee.String() + " is required")
		}
	}

	cb(args)
}
