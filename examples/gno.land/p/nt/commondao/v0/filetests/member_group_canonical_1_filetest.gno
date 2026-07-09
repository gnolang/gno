// PKGPATH: gno.land/r/test/member_group_canonical
package member_group_canonical

import (
	"gno.land/p/nt/commondao/v0"
)

type mockMemberGroup struct{}

func (mockMemberGroup) Seal() commondao.Sealable         { return nil }
func (mockMemberGroup) IsSealed() bool                   { return false }
func (mockMemberGroup) Name() string                     { return "" }
func (mockMemberGroup) Members() commondao.MemberStorage { return nil }
func (mockMemberGroup) SetMeta(any)                      {}
func (mockMemberGroup) GetMeta() any                     { return nil }

var _ commondao.MemberGroup = (*mockMemberGroup)(nil)

func main() {
	g := &mockMemberGroup{}
	println(commondao.IsCanonicalMemberGroup(g))
}

// Output:
// false
