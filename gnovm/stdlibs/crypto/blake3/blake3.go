package blake3

import "lukechampine.com/blake3"

func Sum256(data []byte) [32]byte { return blake3.Sum256(data) }
func Sum512(data []byte) [64]byte { return blake3.Sum512(data) }
