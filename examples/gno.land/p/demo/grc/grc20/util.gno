package grc20

import "std"

const zeroAddress = std.Address("")

func checkIsValidAddress(addr std.Address) error {
	if !addr.IsValid() {
		return ErrInvalidAddress
	}
	return nil
}

func emit(event interface{}) {
	// TODO: setup a pubsub system here?
}
