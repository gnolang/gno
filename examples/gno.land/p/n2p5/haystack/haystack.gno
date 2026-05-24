package haystack

import (
	"encoding/hex"
	"errors"

	"gno.land/p/n2p5/haystack/needle"
	"gno.land/p/nt/avl/v0"
)

var (
	// ErrorNeedleNotFound is returned when a needle is not found in the haystack.
	ErrorNeedleNotFound = errors.New("needle not found")
	// ErrorNeedleLength is returned when a needle is not the correct length.
	ErrorNeedleLength = errors.New("invalid needle length")
	// ErrorHashLength is returned when a needle hash is not the correct length.
	ErrorHashLength = errors.New("invalid hash length")
	// ErrorDuplicateNeedle is returned when a needle already exists in the haystack.
	ErrorDuplicateNeedle = errors.New("needle already exists")
	// ErrorHashMismatch is returned when a needle hash does not match the needle. This should
	// never happen and indicates a critical internal storage error.
	ErrorHashMismatch = errors.New("storage error: hash mismatch")
	// ErrorValueInvalidType is returned when a needle value is not a byte slice. This should
	// never happen and indicates a critical internal storage error.
	ErrorValueInvalidType = errors.New("storage error: invalid value type, expected []byte")
)

const (
	// EncodedHashLength is the length of the hex-encoded needle hash.
	EncodedHashLength = needle.HashLength * 2
	// EncodedPayloadLength is the length of the hex-encoded needle payload.
	EncodedPayloadLength = needle.PayloadLength * 2
	// EncodedNeedleLength is the length of the hex-encoded needle.
	EncodedNeedleLength = EncodedHashLength + EncodedPayloadLength
)

// Haystack is a permissionless, append-only, content-addressed key-value store for fix
// length messages known as needles. A needle is a 192 byte byte slice with a 32 byte
// hash (sha256) and a 160 byte payload.
type Haystack struct{ internal *avl.Tree }

// New creates a new instance of a Haystack key-value store.
func New() *Haystack {
	return &Haystack{
		internal: avl.NewTree(),
	}
}

// Add takes a fixed-length hex-encoded needle bytes and adds it to the haystack key-value
// store. The key is the first 32 bytes of the needle hash (64 bytes hex-encoded) of the
// sha256 sum of the payload. The value is the 160 byte byte slice of the needle payload.
// An error is returned if the needle is found to be invalid.
func (h *Haystack) Add(needleHex string) error {
	if len(needleHex) != EncodedNeedleLength {
		return ErrorNeedleLength
	}
	b, err := hex.DecodeString(needleHex)
	if err != nil {
		return err
	}
	n, err := needle.FromBytes(b)
	if err != nil {
		return err
	}
	if h.internal.Has(needleHex[:EncodedHashLength]) {
		return ErrorDuplicateNeedle
	}
	h.internal.Set(needleHex[:EncodedHashLength], n.Payload())
	return nil
}

// Get takes a hex-encoded needle hash and returns the complete hex-encoded needle bytes
// and an error. Errors covers errors that span from the needle not being found, internal
// storage error inconsistencies, and invalid value types.
func (h *Haystack) Get(hash string) (string, error) {
	if len(hash) != EncodedHashLength {
		return "", ErrorHashLength
	}
	if _, err := hex.DecodeString(hash); err != nil {
		return "", err
	}
	v, ok := h.internal.Get(hash)
	if !ok {
		return "", ErrorNeedleNotFound
	}
	b, ok := v.([]byte)
	if !ok {
		return "", ErrorValueInvalidType
	}
	n, err := needle.New(b)
	if err != nil {
		return "", err
	}
	needleHash := hex.EncodeToString(n.Hash())
	if needleHash != hash {
		return "", ErrorHashMismatch
	}
	return hex.EncodeToString(n.Bytes()), nil
}
