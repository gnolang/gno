// PKGPATH: gno.land/r/test/member_grouping_unsafe
package member_grouping_unsafe

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

func main(cur realm) {
	factory := func(string) commondao.MemberStorage {
		return &mockMemberStorage{}
	}

	// Unsafe member grouping allows a factory returning a non-canonical member storage
	g := commondao.NewUnsafeMemberGrouping(commondao.UseStorageFactory(factory))
	g.Add("name")
	println("ok")
}

// Output:
// ok
