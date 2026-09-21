package needle

import (
	"bytes"
	"crypto/sha256"
	"errors"
)

const (
	// HashLength is the length in bytes of the hash prefix in any message
	HashLength = 32
	// PayloadLength is the length of the remaining bytes of the message.
	PayloadLength = 160
	// NeedleLength is the number of bytes required for a valid needle.
	NeedleLength = HashLength + PayloadLength
)

// Needle is a container for a 160 byte payload
// and a 32 byte sha256 hash of the payload.
type Needle struct {
	hash    [HashLength]byte
	payload [PayloadLength]byte
}

var (
	// ErrorInvalidHash is an error for in invalid hash
	ErrorInvalidHash = errors.New("invalid hash")
	// ErrorByteSliceLength is an error for an invalid byte slice length passed in to New or FromBytes
	ErrorByteSliceLength = errors.New("invalid byte slice length")
)

// New creates a Needle used for submitting a payload to a Haystack sever. It takes a Payload
// byte slice that is 160 bytes in length and returns a reference to a
// Needle and an error. The purpose of this function is to make it
// easy to create a new Needle from a payload. This function handles creating a sha256
// hash of the payload, which is used by the Needle to submit to a haystack server.
func New(p []byte) (*Needle, error) {
	if len(p) != PayloadLength {
		return nil, ErrorByteSliceLength
	}
	var n Needle
	sum := sha256.Sum256(p)
	copy(n.hash[:], sum[:])
	copy(n.payload[:], p)
	return &n, nil
}

// FromBytes is intended convert raw bytes (from UDP or storage) into a Needle.
// It takes a byte slice and expects it to be exactly the length of NeedleLength.
// The byte slice should consist of the first 32 bytes being the sha256 hash of the
// payload and the payload bytes. This function verifies the length of the byte slice,
// copies the bytes into a private [192]byte array, and validates the Needle. It returns
// a reference to a Needle and an error.
func FromBytes(b []byte) (*Needle, error) {
	if len(b) != NeedleLength {
		return nil, ErrorByteSliceLength
	}
	var n Needle
	copy(n.hash[:], b[:HashLength])
	copy(n.payload[:], b[HashLength:])
	if err := n.validate(); err != nil {
		return nil, err
	}
	return &n, nil
}

// Hash returns a copy of the bytes of the sha256 256 hash of the Needle payload.
func (n *Needle) Hash() []byte {
	return n.Bytes()[:HashLength]
}

// Payload returns a byte slice of the Needle payload
func (n *Needle) Payload() []byte {
	return n.Bytes()[HashLength:]
}

// Bytes returns a byte slice of the entire 192 byte hash + payload
func (n *Needle) Bytes() []byte {
	b := make([]byte, NeedleLength)
	copy(b, n.hash[:])
	copy(b[HashLength:], n.payload[:])
	return b
}

// validate checks that a Needle has a valid hash, it returns either nil or an error.
func (n *Needle) validate() error {
	if hash := sha256.Sum256(n.Payload()); !bytes.Equal(n.Hash(), hash[:]) {
		return ErrorInvalidHash
	}
	return nil
}
