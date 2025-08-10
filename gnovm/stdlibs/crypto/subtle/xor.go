package subtle

import (
	"crypto/subtle"
)

func X_xorBytes(dst, x, y []byte) (int, []byte) {
	//XXX: subtle.XORBytes modifies the array of bytes passed as parameter
	// For some reason when using native bindings the array is returned unmodified
	// This was causing this function to behave differently/unexpectedly
	// This hack allows us to also return the modified array
	// originally the function only returns an integer
	return subtle.XORBytes(dst, x, y), dst
}
