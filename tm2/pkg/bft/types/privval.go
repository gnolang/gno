package types

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// PrivValidator defines the functionality of a local Tendermint validator
// that signs votes and proposals, and never double signs.
type PrivValidator interface {
	PubKey() (crypto.PubKey, error)
	SignVote(chainID string, vote *Vote) error
	SignProposal(chainID string, proposal *Proposal) error
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
	pvi, err := pvba[i].PubKey()
	if err != nil {
		panic(err)
	}
	pvj, err := pvba[j].PubKey()
	if err != nil {
		panic(err)
	}

	return pvi.Address().Compare(pvj.Address()) == -1
}

// Swap implements sort.Interface.
func (sba PrivValidatorsByAddress) Swap(i int, j int) {
	it := sba[i]
	sba[i] = sba[j]
	sba[j] = it
}

// mockPV implements PrivValidator without any safety or persistence.
// Only use it for testing.
type mockPV struct {
	signer Signer
}

// mockPV type implements PrivValidator.
var _ PrivValidator = &mockPV{}

// PubKey implements PrivValidator.
func (pv *mockPV) PubKey() (crypto.PubKey, error) {
	return pv.signer.PubKey()
}

// SignVote implements PrivValidator.
func (pv *mockPV) SignVote(chainID string, vote *Vote) error {
	useChainID := chainID
	signBytes := vote.SignBytes(useChainID)
	sig, err := pv.signer.Sign(signBytes)
	if err != nil {
		return err
	}
	vote.Signature = sig
	return nil
}

// SignProposal implements PrivValidator.
func (pv *mockPV) SignProposal(chainID string, proposal *Proposal) error {
	useChainID := chainID
	signBytes := proposal.SignBytes(useChainID)
	sig, err := pv.signer.Sign(signBytes)
	if err != nil {
		return err
	}
	proposal.Signature = sig
	return nil
}

// mockPV type implements fmt.Stringer.
var _ fmt.Stringer = &mockPV{}

// String implements fmt.Stringer.
func (pv *mockPV) String() string {
	pk, _ := pv.signer.PubKey()
	return fmt.Sprintf("MockPV{%v}", pk.Address())
}

// NewMockPVWithPrivKey returns a new MockPV instance.
func NewMockPVWithPrivKey(privKey crypto.PrivKey) *mockPV {
	return &mockPV{newMockSignerWithPrivKey(privKey)}
}

// NewMockPV returns a new MockPV instance.
func NewMockPV() *mockPV {
	return &mockPV{newMockSigner()}
}
