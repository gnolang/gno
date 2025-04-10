package grc721

import "chain"

var zeroAddress = chain.Address("")

func isValidAddress(addr chain.Address) error {
	if !addr.IsValid() {
		return ErrInvalidAddress
	}
	return nil
}

func emit(event any) {
	// TODO: setup a pubsub system here?
}
