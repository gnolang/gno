package sha256

import (
	isha256 "internal/crypto/sha256"
)

const Size = 32

// Sum returns the SHA-256 checksum of the data.
func Sum256(data []byte) [Size]byte {
	return isha256.Sum256(data)
}
