package upgradea

import "std"

const admin = "blahblah"

var (
	counter     int
	nextVersion string
)

func Inc(nb int) {
	if nextVersion != "" {
		panic("contract is locked, please use the new version at r/" + nextVersion)
	}
	counter += nb
}

func Get() int {
	return counter
}

func SetNextVersion(addr string) {
	crossing()
	// assert CallTx call.
	std.AssertOriginCall()
	// assert admin.
	caller := std.PreviousRealm().Address()
	if caller != admin {
		panic("unauthorized")
	}
	nextVersion = addr
}
