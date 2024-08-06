package events

import (
	"std"

	"gno.land/p/demo/ownable/exts/authorizable"
)

var auth *authorizable.Authorizable

func GetOwner() std.Address {
	return auth.Owner()
}

func AddModerator(mod std.Address) {
	auth.AssertCallerIsOwner()

	if err := auth.AddToAuthList(mod); err != nil {
		panic(err)
	}
}
