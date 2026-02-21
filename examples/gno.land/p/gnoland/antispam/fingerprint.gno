package antispam

import (
	"gno.land/p/nt/avl"
)

const (
	// fpShingleSize is the character group size for fingerprinting.
	fpShingleSize = 3

	// fpSignatureSize is the number of hash values kept per fingerprint.
	fpSignatureSize = 32

	// fpDefaultThreshold is the default similarity threshold (0-100).
	fpDefaultThreshold = 50

	// fpMinTextLen is the minimum text length to fingerprint.
	fpMinTextLen = 10

	// fpMaxStoreSize caps the fingerprint store. Oldest entries are evicted.
	fpMaxStoreSize = 500
)

// ---- MinHash fingerprinting

// Fingerprint computes a signature for a text, used for duplicate detection.
// Returns nil for text too short to fingerprint.
func Fingerprint(text string) []uint64 {
	if len(text) < fpMinTextLen {
		return nil
	}

	sig := make([]uint64, fpSignatureSize)
	for i := range sig {
		sig[i] = ^uint64(0) // max uint64
	}

	var w0, w1, w2 rune
	count := 0
	for _, r := range text {
		w0, w1, w2 = w1, w2, r
		count++
		if count < fpShingleSize {
			continue
		}

		baseHash := hashThreeRunes(w0, w1, w2)

		// Derive independent hash functions via SplitMix64 bit mixer.
		for j := 0; j < fpSignatureSize; j++ {
			h := baseHash ^ uint64(j)*0x9e3779b97f4a7c15
			h = (h ^ (h >> 30)) * 0xbf58476d1ce4e5b9
			h = (h ^ (h >> 27)) * 0x94d049bb133111eb
			h = h ^ (h >> 31)

			if h < sig[j] {
				sig[j] = h
			}
		}
	}

	if count < fpShingleSize {
		return nil
	}

	return sig
}

// hashThreeRunes computes a hash for exactly 3 runes.
func hashThreeRunes(a, b, c rune) uint64 {
	h := uint64(14695981039346656037)
	h ^= uint64(byte(a >> 24))
	h *= 1099511628211
	h ^= uint64(byte(a >> 16))
	h *= 1099511628211
	h ^= uint64(byte(a >> 8))
	h *= 1099511628211
	h ^= uint64(byte(a))
	h *= 1099511628211
	h ^= uint64(byte(b >> 24))
	h *= 1099511628211
	h ^= uint64(byte(b >> 16))
	h *= 1099511628211
	h ^= uint64(byte(b >> 8))
	h *= 1099511628211
	h ^= uint64(byte(b))
	h *= 1099511628211
	h ^= uint64(byte(c >> 24))
	h *= 1099511628211
	h ^= uint64(byte(c >> 16))
	h *= 1099511628211
	h ^= uint64(byte(c >> 8))
	h *= 1099511628211
	h ^= uint64(byte(c))
	h *= 1099511628211
	return h
}

// fingerprintSimilarity computes the similarity between two signatures.
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

// ---- FingerprintStore

// FingerprintStore stores fingerprints of known spam content.
// Capped at fpMaxStoreSize entries; the oldest entry is evicted when full.
type FingerprintStore struct {
	tree  *avl.Tree
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

	// Evict the oldest entry if at capacity.
	if fs.tree.Size() >= fpMaxStoreSize {
		fs.tree.Iterate("", "", func(key string, _ interface{}) bool {
			fs.tree.Remove(key)
			return true
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

// intToKey converts an int to a zero-padded string key.
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
