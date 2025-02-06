package types

import (
	"errors"
	"fmt"
	"sort"

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

// SignerByAddress implements sort.Interface for []Signer based on the Address.
type SignerByAddress []Signer

// SignerByAddress type implements sort.Interface.
var _ sort.Interface = SignerByAddress(nil)

// Len implements sort.Interface.
func (sba SignerByAddress) Len() int {
	return len(sba)
}

// Less implements sort.Interface.
func (sba SignerByAddress) Less(i int, j int) bool {
	si, err := sba[i].PubKey()
	if err != nil {
		panic(err)
	}
	sj, err := sba[j].PubKey()
	if err != nil {
		panic(err)
	}

	return si.Address().Compare(sj.Address()) == -1
}

// Swap implements sort.Interface.
func (sba SignerByAddress) Swap(i int, j int) {
	it := sba[i]
	sba[i] = sba[j]
	sba[j] = it
}

// mockSigner implements Signer without persistence. Only use it for testing.
type mockSigner struct {
	privKey crypto.PrivKey
}

// MockSigner type implements Signer.
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

// MockSigner type implements fmt.Stringer.
var _ fmt.Stringer = &mockSigner{}

// String implements fmt.Stringer.
func (ms *mockSigner) String() string {
	pk, _ := ms.PubKey()
	addr := pk.Address()
	return fmt.Sprintf("MockSigner{%v}", addr)
}

// NewMockSigner returns a new mockSigner instance.
func NewMockSigner() Signer {
	return &mockSigner{ed25519.GenPrivKey()}
}

// erroringMockSigner implements Signer that only returns errors. Only use it for testing.
type erroringMockSigner struct{}

// erroringMockSigner type implements Signer.
var _ Signer = &erroringMockSigner{}

var ErrMockSigner = errors.New("erroringMockSigner always returns an error")

// PubKey implements Signer.
func (ems *erroringMockSigner) PubKey() (crypto.PubKey, error) {
	return nil, ErrMockSigner
}

// Sign implements Signer.
func (ems *erroringMockSigner) Sign([]byte) ([]byte, error) {
	return nil, ErrMockSigner
}

// NewErroringMockSigner returns a new erroringMockSigner instance.
func NewErroringMockSigner() Signer {
	return &erroringMockSigner{}
}
