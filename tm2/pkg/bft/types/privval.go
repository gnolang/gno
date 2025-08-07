package types

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// PrivValidator defines the functionality of a local Tendermint validator
// that signs votes and proposals, and never double signs.
type PrivValidator interface {
	PubKey() crypto.PubKey
	SignVote(chainID string, vote *Vote) error
	SignProposal(chainID string, proposal *Proposal) error
	Close() error
}

// PrivValidatorsByAddress implements sort.Interface for []PrivValidator based
// on the Address.
type PrivValidatorsByAddress []PrivValidator

// Len implements sort.Interface.
func (pvba PrivValidatorsByAddress) Len() int {
	return len(pvba)
}

// Less implements sort.Interface.
func (pvba PrivValidatorsByAddress) Less(i int, j int) bool {
	pvi := pvba[i].PubKey()
	pvj := pvba[j].PubKey()

	return pvi.Address().Compare(pvj.Address()) == -1
}

// Swap implements sort.Interface.
func (pvba PrivValidatorsByAddress) Swap(i int, j int) {
	it := pvba[i]
	pvba[i] = pvba[j]
	pvba[j] = it
}

// mockPV implements PrivValidator without any safety or persistence.
// Only use it for testing.
type mockPV struct {
	signer Signer
}

// mockPV type implements PrivValidator.
var _ PrivValidator = &mockPV{}

// PubKey implements PrivValidator.
func (pv *mockPV) PubKey() crypto.PubKey {
	return pv.signer.PubKey()
}

// SignVote implements PrivValidator.
func (pv *mockPV) SignVote(chainID string, vote *Vote) error {
	signBytes := vote.SignBytes(chainID)
	sig, err := pv.signer.Sign(signBytes)
	if err != nil {
		return err
	}
	vote.Signature = sig
	return nil
}

// SignProposal implements PrivValidator.
func (pv *mockPV) SignProposal(chainID string, proposal *Proposal) error {
	signBytes := proposal.SignBytes(chainID)
	sig, err := pv.signer.Sign(signBytes)
	if err != nil {
		return err
	}
	proposal.Signature = sig
	return nil
}

// Close implements PrivValidator.
func (pv *mockPV) Close() error {
	return nil
}

// mockPV type implements fmt.Stringer.
var _ fmt.Stringer = &mockPV{}

// String implements fmt.Stringer.
func (pv *mockPV) String() string {
	return fmt.Sprintf("MockPV{%v}", pv.signer.PubKey().Address())
}

// NewMockPVWithPrivKey returns a new MockPV instance.
func NewMockPVWithPrivKey(privKey crypto.PrivKey) *mockPV {
	return &mockPV{NewMockSignerWithPrivKey(privKey)}
}

// NewMockPV returns a new MockPV instance.
func NewMockPV() *mockPV {
	return &mockPV{NewMockSigner()}
}
