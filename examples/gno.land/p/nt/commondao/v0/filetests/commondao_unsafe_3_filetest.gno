// PKGPATH: gno.land/r/test/commondao_unsafe
package commondao_unsafe

import (
	"gno.land/p/nt/commondao/v0"
)

type mockProposalStorage struct{}

func (mockProposalStorage) Has(uint64) bool                { return false }
func (mockProposalStorage) Get(uint64) *commondao.Proposal { return nil }
func (mockProposalStorage) Add(*commondao.Proposal)        {}
func (mockProposalStorage) Remove(uint64)                  {}
func (mockProposalStorage) Size() int                      { return 0 }

func (mockProposalStorage) Iterate(int, int, bool, func(*commondao.Proposal) bool) bool {
	return false
}

var _ commondao.ProposalStorage = (*mockProposalStorage)(nil)

func main() {
	mock := &mockProposalStorage{}

	// Safe CommonDAO panics with a non-canonical finished proposal storage
	commondao.New(commondao.WithFinishedProposalStorage(mock))
}

// Error:
// CommonDAO finished proposals storage is not canonical
