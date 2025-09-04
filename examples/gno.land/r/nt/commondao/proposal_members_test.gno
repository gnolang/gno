package commondao

import (
	"chain"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	"gno.land/p/moul/addrset"
	"gno.land/p/nt/commondao"
)

var _ commondao.ProposalDefinition = (*membersUpdatePropDefinition)(nil)

func TestMembersPropDefinitionTally(t *testing.T) {
	cases := []struct {
		name    string
		members []chain.Address
		votes   []commondao.Vote
		err     error
		success bool
	}{
		{
			name: "succeed",
			members: []chain.Address{
				"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq",
			},
			votes: []commondao.Vote{
				{
					Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
					Choice:  commondao.ChoiceYes,
				},
				{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceYes,
				},
			},
			success: true,
		},
		{
			name: "fail",
			members: []chain.Address{
				"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq",
			},
			votes: []commondao.Vote{
				{
					Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
					Choice:  commondao.ChoiceNo,
				},
				{
					Address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
					Choice:  commondao.ChoiceNo,
				},
			},
		},
		{
			name: "no quorum",
			members: []chain.Address{
				"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq",
			},
			votes: []commondao.Vote{
				{
					Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
					Choice:  commondao.ChoiceYes,
				},
			},
			err: commondao.ErrNoQuorum,
		},
		{
			name: "succeed with two members",
			members: []chain.Address{
				"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			},
			votes: []commondao.Vote{
				{
					Address: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
					Choice:  commondao.ChoiceYes,
				},
			},
			success: true,
		},
		{
			name: "fail with two members",
			members: []chain.Address{
				"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				p       membersUpdatePropDefinition
				record  commondao.VotingRecord
				members = commondao.NewMemberStorage()
			)

			for _, m := range tc.members {
				members.Add(m)
			}

			for _, v := range tc.votes {
				record.AddVote(v)
			}

			success, err := p.Tally(record.Readonly(), commondao.NewMemberSet(members))

			if tc.err != nil {
				urequire.ErrorIs(t, err, tc.err, "expect an error")
				uassert.False(t, success, "expect tally to fail")
				return
			}

			urequire.NoError(t, err, "expect no error")
			uassert.Equal(t, tc.success, success, "expect tally success to match")
		})
	}
}

func TestMembersPropDefinitionExecute(t *testing.T) {
	cases := []struct {
		name                 string
		dao                  *commondao.CommonDAO
		members, add, remove []chain.Address
		errMsg               string
	}{
		{
			name: "add member",
			dao: commondao.New(
				commondao.WithMember("g1xe2ljac8256rwxxytqddvrjsj2tyv90fvgeaae"),
				commondao.WithMember("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"),
			),
			add: []chain.Address{"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"},
			members: []chain.Address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
				"g1xe2ljac8256rwxxytqddvrjsj2tyv90fvgeaae",
			},
		},
		{
			name: "add multiple members",
			dao: commondao.New(
				commondao.WithMember("g1xe2ljac8256rwxxytqddvrjsj2tyv90fvgeaae"),
			),
			add: []chain.Address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
			},
			members: []chain.Address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
				"g1xe2ljac8256rwxxytqddvrjsj2tyv90fvgeaae",
			},
		},
		{
			name: "remove member",
			dao: commondao.New(
				commondao.WithMember("g1xe2ljac8256rwxxytqddvrjsj2tyv90fvgeaae"),
				commondao.WithMember("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"),
			),
			remove:  []chain.Address{"g1xe2ljac8256rwxxytqddvrjsj2tyv90fvgeaae"},
			members: []chain.Address{"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"},
		},
		{
			name: "remove multiple members",
			dao: commondao.New(
				commondao.WithMember("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
				commondao.WithMember("g1xe2ljac8256rwxxytqddvrjsj2tyv90fvgeaae"),
				commondao.WithMember("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"),
			),
			remove: []chain.Address{
				"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			},
			members: []chain.Address{"g1xe2ljac8256rwxxytqddvrjsj2tyv90fvgeaae"},
		},
		{
			name: "add and remove members",
			dao: commondao.New(
				commondao.WithMember("g1xe2ljac8256rwxxytqddvrjsj2tyv90fvgeaae"),
				commondao.WithMember("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"),
			),
			add:    []chain.Address{"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"},
			remove: []chain.Address{"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"},
			members: []chain.Address{
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g1xe2ljac8256rwxxytqddvrjsj2tyv90fvgeaae",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var add, remove addrset.Set

			for _, addr := range tc.add {
				add.Add(addr)
			}

			for _, addr := range tc.remove {
				remove.Add(addr)
			}

			p := newMembersUpdatePropDefinition(tc.dao, add, remove)

			err := p.Execute(cross)

			if tc.errMsg != "" {
				uassert.ErrorContains(t, err, tc.errMsg, "expect error")
				return
			}

			urequire.NoError(t, err, "expect no error")
			urequire.Equal(t, len(tc.members), tc.dao.Members().Size(), "number of members must match")

			var i int
			tc.dao.Members().IterateByOffset(0, tc.dao.Members().Size(), func(addr chain.Address) bool {
				urequire.Equal(t, tc.members[i], addr, "member address must match")

				i++
				return false
			})
		})
	}
}
