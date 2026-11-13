package merkle

import (
	"github.com/gnolang/gno/tm2/pkg/crypto/tmhash"
)

// TODO: make these have a large predefined capacity
var (
	leafPrefix  = []byte{0}
	innerPrefix = []byte{1}
)

// returns tmhash(0x00 || leaf)
func leafHash(leaf []byte) []byte {
	return tmhash.Sum(append(leafPrefix, leaf...))
}

// leafHashArray returns tmhash(0x00 || leaf), with no heap allocation.
func leafHashArray(leaf []byte) [tmhash.Size]byte {
	h := tmhash.New()
	h.Write(leafPrefix)
	h.Write(leaf)
	var out [tmhash.Size]byte
	// Sum appends, so the empty out[:0] makes it write the digest into out.
	h.Sum(out[:0])
	return out
}

// returns tmhash(0x01 || left || right)
func innerHash(left []byte, right []byte) []byte {
	return tmhash.Sum(append(innerPrefix, append(left, right...)...))
}

// innerHashArray returns tmhash(0x01 || left || right), from the stack.
func innerHashArray(left, right [tmhash.Size]byte) [tmhash.Size]byte {
	var buf [1 + 2*tmhash.Size]byte
	buf[0] = innerPrefix[0]
	copy(buf[1:], left[:])
	copy(buf[1+tmhash.Size:], right[:])
	var out [tmhash.Size]byte
	copy(out[:], tmhash.Sum(buf[:]))
	return out
}
