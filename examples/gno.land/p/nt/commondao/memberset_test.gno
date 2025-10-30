package commondao_test

import (
	"testing"

	"gno.land/p/nt/uassert"

	"gno.land/p/nt/commondao"
)

func TestMemberSetSize(t *testing.T) {
	storage := commondao.NewMemberStorage()
	members := commondao.NewMemberSet(storage)
	uassert.Equal(t, 0, members.Size(), "expect size 0")

	storage.Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	members = commondao.NewMemberSet(storage)
	uassert.Equal(t, 1, members.Size(), "expect size 1")

	storage.Add("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
	members = commondao.NewMemberSet(storage)
	uassert.Equal(t, 2, members.Size(), "expect size 2")
}

func TestMemberSetHas(t *testing.T) {
	storage := commondao.NewMemberStorage()
	storage.Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	members := commondao.NewMemberSet(storage)

	uassert.True(t, members.Has("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"), "expect member is found")
	uassert.False(t, members.Has("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"), "expect member is not found")
}

func TestMemberSetIterateByOffset(t *testing.T) {
	storage := commondao.NewMemberStorage()
	storage.Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	storage.Add("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
	storage.Add("g1w4ek2u3jta047h6lta047h6lta047h6l9huexc")

	commondao.NewMemberSet(storage).IterateByOffset(1, 1, func(addr address) bool {
		uassert.Equal(t, "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", addr.String(), "expect address to match")
		return true
	})
}
