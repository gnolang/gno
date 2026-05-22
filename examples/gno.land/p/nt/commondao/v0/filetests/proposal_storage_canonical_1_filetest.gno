// PKGPATH: gno.land/r/test/proposal_storage_canonical
package proposal_storage_canonical

import (
	"gno.land/p/nt/commondao/v0"
)

type mockProposalStorage struct{}

func (mockProposalStorage) Has(id uint64) bool                { return false }
func (mockProposalStorage) Get(id uint64) *commondao.Proposal { return nil }
func (mockProposalStorage) Add(*commondao.Proposal)           {}
func (mockProposalStorage) Remove(id uint64)                  {}
func (mockProposalStorage) Size() int                         { return 0 }

func (mockProposalStorage) Iterate(int, int, bool, func(*commondao.Proposal) bool) bool {
	return false
}

var _ commondao.ProposalStorage = (*mockProposalStorage)(nil)

func main(cur realm) {
	s := &mockProposalStorage{}
	println(commondao.IsCanonicalProposalStorage(s))
}

// Output:
// false
