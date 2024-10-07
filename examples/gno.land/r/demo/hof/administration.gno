package hof

import "std"

// Exposing the ownable & pausable APIs
// Should not be needed as soon as MsgCall supports calling methods on exported variables

func Pause() error {
	return exhibition.Pausable.Pause()
}

func Unpause() error {
	return exhibition.Pausable.Unpause()
}

func GetOwner() std.Address {
	return owner.Owner()
}

func TransferOwnership(newOwner std.Address) {
	if err := owner.TransferOwnership(newOwner); err != nil {
		panic(err)
	}
}
