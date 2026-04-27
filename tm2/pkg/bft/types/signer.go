package types

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

// Signer is an interface for signing arbitrary byte slices.
type Signer interface {
	PubKey() crypto.PubKey
	Sign([]byte) ([]byte, error)
	Close() error
}

// mockSigner implements Signer without persistence. Only use it for testing.
type mockSigner struct {
	privKey crypto.PrivKey
}

// mockSigner type implements Signer.
var _ Signer = &mockSigner{}

// PubKey implements Signer.
func (ms *mockSigner) PubKey() crypto.PubKey {
	return ms.privKey.PubKey()
}

// Sign implements Signer.
func (ms *mockSigner) Sign(signBytes []byte) ([]byte, error) {
	signature, err := ms.privKey.Sign(signBytes)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// Close implements Signer.
func (ms *mockSigner) Close() error {
	return nil
}

// mockSigner type implements fmt.Stringer.
var _ fmt.Stringer = &mockSigner{}

// String implements fmt.Stringer.
func (ms *mockSigner) String() string {
	return fmt.Sprintf("mockSigner{%v}", ms.PubKey().Address())
}

// NewMockSignerWithPrivKey returns a new mockSigner instance.
func NewMockSignerWithPrivKey(privKey crypto.PrivKey) Signer {
	return &mockSigner{privKey}
}

// NewMockSigner returns a new mockSigner instance.
func NewMockSigner() Signer {
	return &mockSigner{ed25519.GenPrivKey()}
}

// erroringMockSigner implements Signer that only returns error. Only used for testing.
type erroringMockSigner struct {
	privKey crypto.PrivKey
}

// ErrErroringMockSigner is be systematically returned by any call to erroringMockSigner.
var ErrErroringMockSigner = errors.New("erroringMockSigner error")

// erroringMockSigner type implements Signer.
var _ Signer = &erroringMockSigner{}

// PubKey implements Signer.
func (ems *erroringMockSigner) PubKey() crypto.PubKey {
	return ems.privKey.PubKey()
}

// Sign implements Signer.
func (ems *erroringMockSigner) Sign(signBytes []byte) ([]byte, error) {
	return nil, ErrErroringMockSigner
}

// Close implements Signer.
func (ems *erroringMockSigner) Close() error {
	return ErrErroringMockSigner
}

// erroringMockSigner type implements fmt.Stringer.
var _ fmt.Stringer = &erroringMockSigner{}

// String implements fmt.Stringer.
func (ems *erroringMockSigner) String() string {
	return "erroringMockSigner"
}

// NewErroringMockSigner returns a new erroringMockSigner instance.
func NewErroringMockSigner() Signer {
	return &erroringMockSigner{ed25519.GenPrivKey()}
}
