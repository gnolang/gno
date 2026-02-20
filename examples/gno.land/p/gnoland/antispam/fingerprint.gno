package antispam

import (
	"gno.land/p/nt/avl"
)

const (
	// fpShingleSize is the character n-gram size for fingerprinting.
	fpShingleSize = 3

	// fpSignatureSize is the number of min-hash values to keep.
	fpSignatureSize = 64

	// fpDefaultThreshold is the default similarity threshold (0-100).
	fpDefaultThreshold = 50

	// fpMinTextLen is the minimum text length to fingerprint.
	fpMinTextLen = 10

	// fpMaxStoreSize caps the fingerprint store to prevent unbounded O(N)
	// iteration in HasSimilar. Oldest entries are evicted (FIFO).
	fpMaxStoreSize = 500
)

// ---- MinHash fingerprinting

// Fingerprint computes a MinHash signature for a text.
// It extracts character n-grams (shingles), hashes each one directly
// from the rune slice (no string allocation per shingle), and keeps
// the minimum hash value per hash function as the signature.
// Returns nil for text too short to fingerprint.
func Fingerprint(text string) []uint64 {
	if len(text) < fpMinTextLen {
		return nil
	}

	runes := []rune(text)
	if len(runes) < fpShingleSize {
		return nil
	}

	numShingles := len(runes) - fpShingleSize + 1
	if numShingles <= 0 {
		return nil
	}

	sig := make([]uint64, fpSignatureSize)
	for i := range sig {
		sig[i] = ^uint64(0) // max uint64
	}

	for i := 0; i < numShingles; i++ {
		baseHash := hashRunes(runes[i : i+fpShingleSize])

		for j := 0; j < fpSignatureSize; j++ {
			// Different "hash function" by mixing with the index
			h := baseHash ^ uint64(j)*0x9e3779b97f4a7c15
			h = (h ^ (h >> 30)) * 0xbf58476d1ce4e5b9
			h = (h ^ (h >> 27)) * 0x94d049bb133111eb
			h = h ^ (h >> 31)

			if h < sig[j] {
				sig[j] = h
			}
		}
	}

	return sig
}

// hashRunes computes FNV-1a hash directly from a rune slice.
// Avoids string/[]byte allocation per shingle.
// Always hashes 4 bytes per rune (big-endian, fixed layout) to avoid
// branching per byte in the GnoVM interpreter.
func hashRunes(rs []rune) uint64 {
	h := uint64(14695981039346656037)
	for _, r := range rs {
		h ^= uint64(byte(r >> 24))
		h *= 1099511628211
		h ^= uint64(byte(r >> 16))
		h *= 1099511628211
		h ^= uint64(byte(r >> 8))
		h *= 1099511628211
		h ^= uint64(byte(r))
		h *= 1099511628211
	}
	return h
}

// hashString computes a deterministic FNV-1a hash for a string.
func hashString(s string) uint64 {
	h := uint64(14695981039346656037)
	for _, b := range []byte(s) {
		h ^= uint64(b)
		h *= 1099511628211
	}
	return h
}

// fingerprintSimilarity computes the Jaccard similarity between two MinHash signatures.
// Returns a percentage (0-100).
func fingerprintSimilarity(a, b []uint64) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	same := 0
	for i := 0; i < n; i++ {
		if a[i] == b[i] {
			same++
		}
	}
	return same * 100 / n
}

// ---- FingerprintStore (AVL-tree-backed, capped)

// FingerprintStore stores MinHash fingerprints of known spam content.
// Capped at fpMaxStoreSize entries with FIFO eviction.
type FingerprintStore struct {
	tree  *avl.Tree // sequential key -> []uint64
	count int
}

// NewFingerprintStore creates an empty FingerprintStore.
func NewFingerprintStore() *FingerprintStore {
	return &FingerprintStore{
		tree: avl.NewTree(),
	}
}

// Add computes and stores the fingerprint of a text.
// If the store is at capacity, the oldest entry is evicted.
func (fs *FingerprintStore) Add(content string) {
	fp := Fingerprint(content)
	if fp == nil {
		return
	}

	// Evict oldest if at capacity
	if fs.tree.Size() >= fpMaxStoreSize {
		// Find and remove the first (oldest) key
		fs.tree.Iterate("", "", func(key string, _ interface{}) bool {
			fs.tree.Remove(key)
			return true // stop after first
		})
	}

	key := intToKey(fs.count)
	fs.tree.Set(key, fp)
	fs.count++
}

// HasSimilar checks if any stored fingerprint is similar to the given content.
// Threshold is a percentage (0-100): 80 means 80% similarity required.
func (fs *FingerprintStore) HasSimilar(content string, threshold int) bool {
	fp := Fingerprint(content)
	if fp == nil {
		return false
	}

	found := false
	fs.tree.Iterate("", "", func(key string, value interface{}) bool {
		stored := value.([]uint64)
		sim := fingerprintSimilarity(fp, stored)
		if sim >= threshold {
			found = true
			return true // stop
		}
		return false
	})
	return found
}

// Size returns the number of stored fingerprints.
func (fs *FingerprintStore) Size() int {
	return fs.tree.Size()
}

// intToKey converts an int to a zero-padded string key for AVL tree ordering.
func intToKey(n int) string {
	var buf [10]byte
	for i := 9; i >= 0; i-- {
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[:])
}

// ---- Scoring

// ScoreFingerprint checks if content is a near-duplicate of known spam.
func ScoreFingerprint(content string, fps *FingerprintStore) (int, string) {
	if fps == nil || fps.Size() == 0 {
		return 0, ""
	}
	if fps.HasSimilar(content, fpDefaultThreshold) {
		return WeightNearDuplicate, "NEAR_DUPLICATE"
	}
	return 0, ""
}
