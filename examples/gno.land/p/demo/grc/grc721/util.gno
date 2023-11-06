package grc721

import (
	"std"
)

var zeroAddress = std.Address("")

func isValidAddress(addr std.Address) error {
	if !addr.IsValid() {
		return ErrInvalidAddress
	}
	return nil
}

func emit(event interface{}) {
	// TODO: setup a pubsub system here?
}
