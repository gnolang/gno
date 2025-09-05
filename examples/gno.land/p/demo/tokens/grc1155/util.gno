package grc1155

import (
	"std"
)

const zeroAddress std.Address = ""

func isValidAddress(addr std.Address) bool {
	if !addr.IsValid() {
		return false
	}
	return true
}

func emit(event any) {
	// TODO: setup a pubsub system here?
}
