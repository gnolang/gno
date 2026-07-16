package grc721

const (
	MaxNameLen   = 64
	MaxSymbolLen = 11
)

var zeroAddress = address("")

func isValidAddress(addr address) error {
	if !addr.IsValid() {
		return ErrInvalidAddress
	}
	return nil
}

func validName(name string) bool {
	if name == "" || len(name) > MaxNameLen {
		return false
	}
	for _, c := range name {
		if c < 0x20 || c == 0x7f {
			return false
		}
	}
	return true
}

// symbol is emitted as part of the event identifier, so bounding it
// keeps events under MaxEventAttrLen.
func validSymbol(symbol string) bool {
	if symbol == "" || len(symbol) > MaxSymbolLen {
		return false
	}
	for _, c := range symbol {
		if !isAlnum(c) && c != '_' && c != '-' {
			return false
		}
	}
	return true
}

func isAlnum(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func emit(event any) {
	// TODO: setup a pubsub system here?
}
