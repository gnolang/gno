package cid

import (
	"crypto/keccak256"
	"crypto/sha256"
	"errors"
)

// Hash function codes Verify supports (multicodec table). Multihashes with
// other codes still parse.
const (
	Identity  uint64 = 0x00 // the digest is the content itself
	SHA256    uint64 = 0x12 // sha2-256
	Keccak256 uint64 = 0x1b // keccak-256
)

// maxDigestLen caps every digest. Only identity digests, which inline the
// content, come near it.
const maxDigestLen = 128

var (
	// ErrDigestMismatch is returned by Verify when a block does not hash to
	// the CID's digest.
	ErrDigestMismatch = errors.New("cid: block does not match the CID's digest")
	// ErrUnsupportedHash is returned by Verify for a hash function it does
	// not implement.
	ErrUnsupportedHash = errors.New("cid: unsupported hash function")

	errMultihash = errors.New("cid: invalid multihash")
)

// Multihash is a self-describing digest: a hash function code and the
// digest it produced. Like CID, it holds its binary form and is immutable.
type Multihash struct{ b string }

// NewMultihash returns the multihash of a digest computed with the hash
// function code. sha2-256 and keccak-256 digests must be 32 bytes, and no
// digest may exceed 128 bytes.
func NewMultihash(code uint64, digest []byte) (Multihash, error) {
	if code > maxVarint || !validDigestLen(code, uint64(len(digest))) {
		return Multihash{}, errMultihash
	}
	b := appendUvarint(nil, code)
	b = appendUvarint(b, uint64(len(digest)))
	return Multihash{string(append(b, digest...))}, nil
}

// multihashLen returns the length of the valid multihash at the start of s.
func multihashLen(s string) (int, error) {
	code, n, err := uvarint(s)
	if err != nil {
		return 0, err
	}
	size, m, err := uvarint(s[n:])
	if err != nil {
		return 0, err
	}
	if !validDigestLen(code, size) || uint64(len(s)-n-m) < size {
		return 0, errMultihash
	}
	return n + m + int(size), nil
}

func validDigestLen(code, size uint64) bool {
	switch code {
	case SHA256, Keccak256:
		return size == 32
	}
	return size <= maxDigestLen
}

// split returns the hash function code and the digest.
func (m Multihash) split() (uint64, string) {
	code, n, _ := uvarint(m.b)
	_, k, _ := uvarint(m.b[n:])
	return code, m.b[n+k:]
}

// Code returns the hash function code.
func (m Multihash) Code() uint64 {
	code, _ := m.split()
	return code
}

// Digest returns a copy of the digest.
func (m Multihash) Digest() []byte {
	_, digest := m.split()
	return []byte(digest)
}

// Bytes returns a copy of the binary multihash.
func (m Multihash) Bytes() []byte {
	return []byte(m.b)
}

// verify reports whether block hashes to m.
func (m Multihash) verify(block []byte) error {
	code, digest := m.split()

	var sum string
	switch code {
	case SHA256:
		h := sha256.Sum256(block)
		sum = string(h[:])
	case Keccak256:
		h := keccak256.Sum256(block)
		sum = string(h[:])
	case Identity:
		if len(block) != len(digest) {
			return ErrDigestMismatch
		}
		sum = string(block)
	default:
		return ErrUnsupportedHash
	}

	if sum != digest {
		return ErrDigestMismatch
	}
	return nil
}
