package testutils

import (
	"slices"

	"github.com/gnolang/gno/tm2/pkg/random"
)

// Contract: !bytes.Equal(input, output) && len(input) >= len(output)
// TODO: keep output size the same; search all usage first.
func MutateByteSlice(bytez []byte) []byte {
	// If bytez is empty, panic
	if len(bytez) == 0 {
		panic("Cannot mutate an empty bytez")
	}

	// Copy bytez
	mBytez := make([]byte, len(bytez))
	copy(mBytez, bytez)
	bytez = mBytez

	// Try a random mutation
	switch random.RandInt() % 2 {
	case 0: // Mutate a single byte
		bytez[random.RandInt()%len(bytez)] += byte(random.RandInt()%255 + 1)
	case 1: // Remove an arbitrary byte
		pos := random.RandInt() % len(bytez)
		bytez = slices.Delete(bytez, pos, pos+1)
	}
	return bytez
}
