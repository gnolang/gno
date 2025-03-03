package dao_voting_group

import (
	"testing"

	dao_interfaces "gno.land/p/teritori/dao_interfaces"
)

func TestVotingGroup(t *testing.T) {
	v := NewVotingGroup()
	var i dao_interfaces.IVotingModule
	i = v

	{
		got := i.TotalPowerAtHeight(0, []string{})
		expected := uint64(0)
		if got != expected {
			t.Fatalf("expected %s, got %s.", expected, got)
		}
	}

	{
		conf := v.ConfigJSON()
		expected := `{"totalPower":"0","members":"0"}`
		if conf != expected {
			t.Fatalf("expected %s, got %s.", expected, conf)
		}
	}
}
