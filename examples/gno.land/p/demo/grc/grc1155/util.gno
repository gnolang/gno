package grc1155

import (
	"std"
)

const zeroAddress std.Address = ""

func isValidAddress(addr std.Address) bool {
	if addr.String() == "" {
		return false
	}
	return true
}

func emit(event interface{}) {
	// TODO: setup a pubsub system here?
}
