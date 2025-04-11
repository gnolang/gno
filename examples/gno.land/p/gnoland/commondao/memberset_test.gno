package commondao

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/moul/addrset"
)

func TestMemberSetSize(t *testing.T) {
	var set addrset.Set
	members := NewMemberSet(&set)
	uassert.Equal(t, 0, members.Size(), "expect size 0")

	set.Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	members = NewMemberSet(&set)
	uassert.Equal(t, 1, members.Size(), "expect size 1")

	set.Add("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
	members = NewMemberSet(&set)
	uassert.Equal(t, 2, members.Size(), "expect size 2")
}

func TestMemberSetHas(t *testing.T) {
	var set addrset.Set
	set.Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	members := NewMemberSet(&set)

	uassert.True(t, members.Has("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"), "expect member is found")
	uassert.False(t, members.Has("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"), "expect member is not found")
}

func TestMemberSetIterateByOffset(t *testing.T) {
	var set addrset.Set
	set.Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	set.Add("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
	set.Add("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc")

	NewMemberSet(&set).IterateByOffset(1, 1, func(addr std.Address) bool {
		uassert.Equal(t, "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", addr.String(), "expect address to match")
		return true
	})
}
