package blake3

import "lukechampine.com/blake3"

func X_sum256(data []byte) [32]byte {
	return blake3.Sum256(data)
}
