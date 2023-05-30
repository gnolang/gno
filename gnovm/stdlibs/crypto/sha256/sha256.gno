package sha256

const Size = 32

// Sum returns the SHA-256 checksum of the data.
func Sum256(data []byte) [Size]byte
