// PKGPATH: gno.land/r/test/member_grouping_canonical
package member_grouping_canonical

import (
	"gno.land/p/nt/commondao/v0"
)

type mockMemberGrouping struct{}

func (mockMemberGrouping) Seal() commondao.Sealable { return nil }
func (mockMemberGrouping) IsSealed() bool           { return false }
func (mockMemberGrouping) Size() int                { return 0 }
func (mockMemberGrouping) Has(string) bool          { return false }
func (mockMemberGrouping) Delete(string) error      { return nil }

func (mockMemberGrouping) Add(string) (commondao.MemberGroup, error) {
	return nil, nil
}

func (mockMemberGrouping) Get(string) (commondao.MemberGroup, bool) {
	return nil, false
}

func (mockMemberGrouping) IterateByOffset(int, int, commondao.MemberGroupIterFn) bool {
	return false
}

var _ commondao.MemberGrouping = (*mockMemberGrouping)(nil)

func main(cur realm) {
	g := &mockMemberGrouping{}
	println(commondao.IsCanonicalMemberGrouping(g))
}

// Output:
// false
