// Package internal exposes internal functions used within db packages.
package internal

// NonNilBytes ensures that bz is a non-nil byte slice (ie. []byte{}).
//
// We defensively turn nil keys or values into []byte{} for
// most operations.
func NonNilBytes(bz []byte) []byte {
	if bz == nil {
		return []byte{}
	}
	return bz
}
