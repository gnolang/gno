package grc1155

const zeroAddress address = ""

func isValidAddress(addr address) bool {
	if !addr.IsValid() {
		return false
	}
	return true
}

func emit(event any) {
	// TODO: setup a pubsub system here?
}
