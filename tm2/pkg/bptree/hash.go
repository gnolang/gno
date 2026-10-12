package bptree

import (
	"crypto/sha256"
	"encoding/binary"
)

// HashLeafSlot computes the hash for an occupied leaf slot:
//
//	SHA256(0x00 || varint(len(key)) || key || varint(32) || SHA256(value))
//
// This matches ICS23 LeafOp with Prefix=0x00, PrehashValue=SHA256, Length=VAR_PROTO.
func HashLeafSlot(key, value []byte) Hash {
	valueHash := sha256.Sum256(value)
	return HashLeafSlotFromValueHash(key, valueHash)
}

// HashLeafSlotFromValueHash computes the leaf slot hash from a pre-computed value hash.
//
//	SHA256(0x00 || varint(len(key)) || key || 0x20 || valueHash)
func HashLeafSlotFromValueHash(key []byte, valueHash Hash) Hash {
	h := sha256.New()
	h.Write([]byte{DomainLeaf})
	var vbuf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(vbuf[:], uint64(len(key)))
	h.Write(vbuf[:n])
	h.Write(key)
	h.Write([]byte{0x20})
	h.Write(valueHash[:])
	var result Hash
	h.Sum(result[:0])
	return result
}

// HashInner computes an inner mini-merkle node hash:
//
//	SHA256(0x01 || left || right)
//
// If both left and right are the sentinel, returns the sentinel
// (short-circuit rule for ICS23 EmptyChild compatibility).
func HashInner(left, right Hash) Hash {
	if left == sentinelHash && right == sentinelHash {
		return sentinelHash
	}
	var buf [1 + HashSize + HashSize]byte
	buf[0] = DomainInner
	copy(buf[1:], left[:])
	copy(buf[1+HashSize:], right[:])
	return sha256.Sum256(buf[:])
}
