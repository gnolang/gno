package mock

import (
	"bytes"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/random"
)

// -------------------------------------
// These are fake crypto implementations, useful for testing.

var _ crypto.PrivKey = PrivKeyMock{}

// PrivKeyMock implements crypto.PrivKey.
type PrivKeyMock []byte

// Bytes marshals the privkey using amino encoding w/ type information.
func (privKey PrivKeyMock) Bytes() []byte {
	return amino.MustMarshalAny(privKey)
}

// Make a fake signature.  Its length is variable.
func (privKey PrivKeyMock) Sign(msg []byte) ([]byte, error) {
	sigBytes := fmt.Appendf(nil, "signature-for-%X-by-%X", msg, []byte(privKey))
	return sigBytes, nil
}

func (privKey PrivKeyMock) PubKey() crypto.PubKey {
	return PubKeyMock(privKey)
}

func (privKey PrivKeyMock) Equals(other crypto.PrivKey) bool {
	if otherMock, ok := other.(PrivKeyMock); ok {
		return bytes.Equal(privKey, otherMock)
	} else {
		return false
	}
}

func GenPrivKey() PrivKeyMock {
	randstr := random.RandStr(12)
	return []byte(randstr)
}

// -------------------------------------

var _ crypto.PubKey = PubKeyMock{}

type PubKeyMock []byte

// Returns address w/ pubkey as suffix (for debugging).
func (pubKey PubKeyMock) Address() crypto.Address {
	if len(pubKey) > 20 {
		panic("PubKeyMock cannot have pubkey greater than 20 bytes")
	}
	addr := crypto.Address{}
	copy(addr[20-len(pubKey):], pubKey)
	return addr
}

// Bytes marshals the PubKey using amino encoding.
func (pubKey PubKeyMock) Bytes() []byte {
	return amino.MustMarshalAny(pubKey)
}

func (pubKey PubKeyMock) VerifyBytes(msg []byte, sig []byte) bool {
	sigBytes := fmt.Appendf(nil, "signature-for-%X-by-%X", msg, []byte(pubKey))
	return bytes.Equal(sig, sigBytes)
}

func (pubKey PubKeyMock) String() string {
	return fmt.Sprintf("PubKeyMock{%X}", ([]byte(pubKey))[:])
}

func (pubKey PubKeyMock) Equals(other crypto.PubKey) bool {
	if otherMock, ok := other.(PubKeyMock); ok {
		return bytes.Equal(pubKey[:], otherMock[:])
	} else {
		return false
	}
}
