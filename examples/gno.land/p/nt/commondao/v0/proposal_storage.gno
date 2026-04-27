package commondao

import (
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/seqid/v0"
)

// ProposalStorage defines an interface for proposal storages.
type ProposalStorage interface {
	// Has checks if a proposal exists.
	Has(id uint64) bool

	// Get returns a proposal or nil when proposal doesn't exist.
	Get(id uint64) *Proposal

	// Add adds a proposal to the storage.
	Add(*Proposal)

	// Remove removes a proposal from the storage.
	Remove(id uint64)

	// Size returns the number of proposals that the storage contains.
	Size() int

	// Iterate iterates proposals.
	Iterate(offset, count int, reverse bool, fn func(*Proposal) bool) bool
}

// NewProposalStorage creates a new proposal storage.
func NewProposalStorage() ProposalStorage {
	return &proposalStorage{avl.NewTree()}
}

type proposalStorage struct {
	storage *avl.Tree // string(proposal ID) -> *Proposal
}

// Has checks if a proposal exists.
func (s proposalStorage) Has(id uint64) bool {
	return s.storage.Has(makeProposalKey(id))
}

// Get returns a proposal or nil when proposal doesn't exist.
func (s proposalStorage) Get(id uint64) *Proposal {
	if v, found := s.storage.Get(makeProposalKey(id)); found {
		return v.(*Proposal)
	}
	return nil
}

// Add adds a proposal to the storage.
func (s *proposalStorage) Add(p *Proposal) {
	if p == nil {
		return
	}

	s.storage.Set(makeProposalKey(p.ID()), p)
}

// Remove removes a proposal from the storage.
func (s *proposalStorage) Remove(id uint64) {
	s.storage.Remove(makeProposalKey(id))
}

// Size returns the number of proposals that the storage contains.
func (s proposalStorage) Size() int {
	return s.storage.Size()
}

// Iterate iterates proposals.
func (s proposalStorage) Iterate(offset, count int, reverse bool, fn func(*Proposal) bool) bool {
	if fn == nil {
		return false
	}

	cb := func(_ string, v any) bool { return fn(v.(*Proposal)) }

	if reverse {
		return s.storage.ReverseIterateByOffset(offset, count, cb)
	}
	return s.storage.IterateByOffset(offset, count, cb)
}

func makeProposalKey(id uint64) string {
	return seqid.ID(id).String()
}
