package types

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

// Signer is an interface for signing arbitrary byte slices.
// All methods can return errors because the Signer might be a remote server
// and the connection to it can fail.
type Signer interface {
	PubKey() (crypto.PubKey, error)
	Sign([]byte) ([]byte, error)
}

// mockSigner implements Signer without persistence. Only use it for testing.
type mockSigner struct {
	privKey crypto.PrivKey
}

// mockSigner type implements Signer.
var _ Signer = &mockSigner{}

// PubKey implements Signer.
func (ms *mockSigner) PubKey() (crypto.PubKey, error) {
	return ms.privKey.PubKey(), nil
}

// Sign implements Signer.
func (ms *mockSigner) Sign(signBytes []byte) ([]byte, error) {
	signature, err := ms.privKey.Sign(signBytes)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// mockSigner type implements fmt.Stringer.
var _ fmt.Stringer = &mockSigner{}

// String implements fmt.Stringer.
func (ms *mockSigner) String() string {
	pk, _ := ms.PubKey()
	return fmt.Sprintf("mockSigner{%v}", pk.Address())
}

// newMockSignerWithPrivKey returns a new mockSigner instance.
func newMockSignerWithPrivKey(privKey crypto.PrivKey) Signer {
	return &mockSigner{privKey}
}

// newMockSigner returns a new mockSigner instance.
func newMockSigner() Signer {
	return &mockSigner{ed25519.GenPrivKey()}
}
