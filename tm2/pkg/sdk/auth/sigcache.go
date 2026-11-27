package auth

import (
	"crypto/sha256"
	"encoding/binary"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// SigVerifyCache remembers signature verifications that succeeded, so that a
// transaction checked again on every mempool recheck does not pay for the same
// public-key verification each time.
//
// It is node local and consulted only in CheckTx (which includes recheck).
// DeliverTx always verifies, so the cache cannot change what a block commits.
// Gas is charged before the lookup, so a hit changes time, never gas.
//
// The key is sha256 over every input of the verification: the public key, the
// exact sign bytes (chain id, account number, current sequence, fee, msgs,
// memo) and the signature. A hit therefore means "this exact call already
// returned true", and nothing narrower. Keying on the signature alone would
// accept a replayed tx with a stale sequence, a signature transplanted onto a
// different body, or a tx from another chain.
//
// The cache is bounded: when full, the oldest entry is evicted (FIFO). An
// evicted entry only costs a full verification on its next check.
type SigVerifyCache struct {
	mu   sync.Mutex
	set  map[[sha256.Size]byte]struct{}
	ring [][sha256.Size]byte // insertion order, for eviction
	next int                 // ring slot to overwrite once full
}

// NewSigVerifyCache returns a cache holding at most size entries. It panics if
// size is not positive.
func NewSigVerifyCache(size int) *SigVerifyCache {
	if size <= 0 {
		panic("SigVerifyCache size must be positive")
	}
	return &SigVerifyCache{
		set:  make(map[[sha256.Size]byte]struct{}, size),
		ring: make([][sha256.Size]byte, 0, size),
	}
}

// Len returns the number of cached verifications.
func (c *SigVerifyCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.set)
}

func (c *SigVerifyCache) has(k [sha256.Size]byte) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.set[k]
	return ok
}

func (c *SigVerifyCache) add(k [sha256.Size]byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.set[k]; ok {
		return
	}
	if len(c.ring) < cap(c.ring) {
		c.ring = append(c.ring, k)
	} else {
		delete(c.set, c.ring[c.next])
		c.ring[c.next] = k
		c.next = (c.next + 1) % len(c.ring)
	}
	c.set[k] = struct{}{}
}

// sigVerifyKey hashes the full input of pubKey.VerifyBytes(signBytes, sig).
// Each part is length-prefixed so that no two distinct inputs share a key.
func sigVerifyKey(pubKey crypto.PubKey, signBytes, sig []byte) [sha256.Size]byte {
	h := sha256.New()
	var n [8]byte
	for _, part := range [][]byte{pubKey.Bytes(), signBytes, sig} {
		binary.BigEndian.PutUint64(n[:], uint64(len(part)))
		h.Write(n[:])
		h.Write(part)
	}
	var k [sha256.Size]byte
	h.Sum(k[:0])
	return k
}
