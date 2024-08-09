package events

import (
	"std"

	"gno.land/p/demo/ownable/exts/authorizable"
)

var (
	su   = std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5") // @leohhhn
	auth = authorizable.NewAuthorizableWithAddress(su)
)

func GetOwner() std.Address {
	return auth.Owner()
}

func AddModerator(mod std.Address) {
	auth.AssertCallerIsOwner()

	if err := auth.AddToAuthList(mod); err != nil {
		panic(err)
	}
}
