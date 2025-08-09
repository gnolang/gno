package subtle

import (
	"crypto/subtle"
)

func X_xorBytes(dst, x, y []byte) (int, []byte) {
	return subtle.XORBytes(dst, x, y), dst
}
