package commondao

import (
	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/seqid/v0"
)

// newProposalStorage creates a new proposal storage.
func newProposalStorage() *proposalStorage {
	return &proposalStorage{bptree.NewBPTree32()}
}

// proposalStorage stores proposals indexed by ID.
type proposalStorage struct {
	storage *bptree.BPTree // string(proposal ID) -> *Proposal
}

// Get returns a proposal or nil when proposal doesn't exist.
func (s proposalStorage) Get(id uint64) *Proposal {
	if v := s.storage.Get(makeProposalKey(id)); v != nil {
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
