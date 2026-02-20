package antispam

import (
	"gno.land/p/nt/avl"
)

const (
	WeightNearDuplicate = 4

	// fpShingleSize is the character n-gram size for fingerprinting.
	fpShingleSize = 3

	// fpSignatureSize is the number of min-hash values to keep.
	fpSignatureSize = 64

	// fpDefaultThreshold is the default similarity threshold (0-100).
	fpDefaultThreshold = 50

	// fpMinTextLen is the minimum text length to fingerprint.
	fpMinTextLen = 10
)

// ---- MinHash fingerprinting

// Fingerprint computes a MinHash signature for a text.
// It extracts character n-grams (shingles), hashes each one,
// and keeps the K smallest hash values as the signature.
// Returns nil for text too short to fingerprint.
func Fingerprint(text string) []uint64 {
	if len(text) < fpMinTextLen {
		return nil
	}

	runes := []rune(text)
	if len(runes) < fpShingleSize {
		return nil
	}

	// Compute shingle hashes
	numShingles := len(runes) - fpShingleSize + 1
	if numShingles <= 0 {
		return nil
	}

	// Use multiple hash functions (simulated via different seeds)
	// For each hash function, keep the minimum hash value across all shingles.
	sig := make([]uint64, fpSignatureSize)
	for i := range sig {
		sig[i] = ^uint64(0) // max uint64
	}

	for i := 0; i < numShingles; i++ {
		shingle := string(runes[i : i+fpShingleSize])
		baseHash := hashString(shingle)

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

// hashString computes a deterministic hash for a string.
// Uses FNV-1a algorithm.
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

// ---- FingerprintStore (AVL-tree-backed)

// FingerprintStore stores MinHash fingerprints of known spam content.
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
func (fs *FingerprintStore) Add(content string) {
	fp := Fingerprint(content)
	if fp == nil {
		return
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
	return fs.count
}

// intToKey converts an int to a zero-padded string key for AVL tree ordering.
func intToKey(n int) string {
	if n <= 0 {
		return "0000000000"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	for len(s) < 10 {
		s = "0" + s
	}
	return s
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
