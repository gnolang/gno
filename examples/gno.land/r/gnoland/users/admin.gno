package users

import "std"

// Exposing the ownable & pausable APIs
// Should not be needed as soon as MsgCall supports calling methods on exported variables

func Pause() error {
	return pausable.Pause()
}

func Unpause() error {
	return pausable.Unpause()
}

func GetOwner() std.Address {
	return admin.Owner()
}

func TransferOwnership(newOwner std.Address) {
	if err := admin.TransferOwnership(newOwner); err != nil {
		panic(err)
	}
}
