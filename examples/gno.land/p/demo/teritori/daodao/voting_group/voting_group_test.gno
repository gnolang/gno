package dao_voting_group

import (
	"std"
	"testing"

	dao_interfaces "gno.land/p/demo/teritori/daodao/interfaces"
	"gno.land/r/demo/groups"
)

func TestVotingGroup(t *testing.T) {
	g := groups.CreateGroup("test_voting_group")
	v := NewVotingGroup(g)
	var i dao_interfaces.IVotingModule
	i = v
	got := i.TotalPowerAtHeight(0)
	expected := uint64(0)
	if got != expected {
		t.Fatalf("expected %q, got %q.", expected, got)
	}
}
