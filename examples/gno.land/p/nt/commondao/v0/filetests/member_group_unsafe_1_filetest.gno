// PKGPATH: gno.land/r/test/member_group_unsafe
package member_group_unsafe

import (
	"gno.land/p/nt/commondao/v0"
)

type mockMemberStorage struct{}

func (mockMemberStorage) Seal() commondao.Sealable           { return nil }
func (mockMemberStorage) IsSealed() bool                     { return false }
func (mockMemberStorage) Size() int                          { return 0 }
func (mockMemberStorage) Has(address) bool                   { return false }
func (mockMemberStorage) Add(address) bool                   { return false }
func (mockMemberStorage) Remove(address) bool                { return false }
func (mockMemberStorage) Grouping() commondao.MemberGrouping { return nil }

func (mockMemberStorage) IterateByOffset(int, int, commondao.MemberIterFn) bool {
	return false
}

var _ commondao.MemberStorage = (*mockMemberStorage)(nil)

func main() {
	mock := &mockMemberStorage{}

	// Safe member group panics with a non-canonical member storage
	commondao.NewMemberGroup("name", mock)
	println("ok")
}

// Error:
// member storage for group is not canonical
