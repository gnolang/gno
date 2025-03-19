package halt

import (
	"std"
)

// getPrevRealm returns object of the previous realm.
func getPrevRealm() std.Realm {
	return std.PreviousRealm()
}

// getPrevAddr returns the address of the previous realm.
func getPrevAddr() std.Address {
	return std.PreviousRealm().Address()
}

// getPrevAsString returns the address and package path of the previous realm.
func getPrevAsString() (string, string) {
	prev := getPrevRealm()
	return prev.Address().String(), prev.PkgPath()
}
