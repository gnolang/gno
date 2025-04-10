package grc1155

import "chain"

const zeroAddress chain.Address = ""

func isValidAddress(addr chain.Address) bool {
	if !addr.IsValid() {
		return false
	}
	return true
}

func emit(event any) {
	// TODO: setup a pubsub system here?
}
