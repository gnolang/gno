package gnousd

import "std"

func checkIsValidAddress(addr std.Address) {
	if addr.String() == "" {
		panic("invalid address")
	}
}

func emit(event interface{}) {
	// TODO: should we do something there?
	// noop
}

const zeroAddress = std.Address("")

const reserveAddr = std.Address("reserve") // fake address (could it be a problem?)

func getCurrentHeight() uint64 {
	return uint64(std.GetHeight())
}
