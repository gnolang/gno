package gnft

import (
	"std"

	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ufmt"
)

// getPrevAsString returns the address and package path of the previous realm.
func getPrevAsString() (string, string) {
	prev := std.PreviousRealm()
	return prev.Address().String(), prev.PkgPath()
}

// getPrevAddr returns the address of the previous realm.
func getPrevAddr() std.Address {
	return std.PreviousRealm().Address()
}

// tid converts uint64 to grc721.TokenID.
//
// Input:
//   - id: the uint64 to convert
//
// Output:
//   - grc721.TokenID: the converted token ID
func tid(id uint64) grc721.TokenID {
	return grc721.TokenID(ufmt.Sprintf("%d", id))
}

// Exists checks if a token ID exists.
//
// Input:
//   - tid: the token ID to check
//
// Output:
//   - bool: true if the token ID exists, false otherwise
func Exists(tid grc721.TokenID) bool {
	_, err := nft.OwnerOf(tid)
	if err != nil {
		return false
	}

	return true
}
