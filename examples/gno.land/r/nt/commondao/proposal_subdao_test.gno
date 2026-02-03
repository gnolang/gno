package commondao

import (
	"testing"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/commondao"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
)

var _ commondao.ProposalDefinition = (*subDAOPropDefinition)(nil)

func TestSubDAOPropDefinitionTally(t *testing.T) {
	cases := []struct {
		name    string
		members []address
		votes   []commondao.Vote
		err     error
		success bool
	}{
		{
			name: "succeed",
			members: []address{
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
				{
					Address: "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq",
					Choice:  commondao.ChoiceYes,
				},
			},
			success: true,
		},
		{
			name: "fail",
			members: []address{
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
				{
					Address: "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq",
					Choice:  commondao.ChoiceNo,
				},
			},
		},
		{
			name: "fail with no votes",
			members: []address{
				"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq",
			},
			err: commondao.ErrNoQuorum,
		},
		{
			name: "no quorum",
			members: []address{
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
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				p       subDAOPropDefinition
				record  commondao.VotingRecord
				members = commondao.NewMemberStorage()
			)

			for _, m := range tc.members {
				members.Add(m)
			}

			for _, v := range tc.votes {
				record.AddVote(v)
			}

			ctx := commondao.MustNewVotingContext(&record, members)

			success, err := p.Tally(ctx)

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

func TestSubDAOPropDefinitionExecute(t *testing.T) {
	// Create a parent DAO
	parentDAO := createDAO("Foo", "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", nil)
	urequire.Equal(t, 0, parentDAO.Children().Len(), "expected no initial SubDAOs")

	// Initialize some members for the SubDAO
	var members addrset.Set
	members.Add("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	members.Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")

	// Create the proposal and execute it
	name := "Bar"
	p := newSubDAOPropDefinition(parentDAO, name, &members)
	err := p.Execute(cross)
	urequire.NoError(t, err, "expected no execution error")

	// Expect SubDAO to be created
	urequire.Equal(t, 1, parentDAO.Children().Len(), "expected SubDAO to be created")
	subDAO := parentDAO.Children().Get(0).(*commondao.CommonDAO)

	// Make sure SubDAO is indexed properly
	subDAO = getDAO(subDAO.ID())
	urequire.True(t, subDAO != nil, "expected SubDAO to be indexed by the realm")

	// Make sure SubDAO values are the right ones
	uassert.Equal(t, name, subDAO.Name(), "expected SubDAO name to match")
	uassert.Equal(t, 2, subDAO.Members().Size(), "expected SubDAO to have two members")

	var i int
	subDAO.Members().IterateByOffset(0, subDAO.Members().Size(), func(addr address) bool {
		s, _ := members.Tree().GetByIndex(i)
		urequire.Equal(t, address(s), addr, "expected member to be found")
		i++
		return false
	})
}

func TestDissolvePropDefinitionTally(t *testing.T) {
	cases := []struct {
		name    string
		members []address
		votes   []commondao.Vote
		err     error
		success bool
	}{
		{
			name: "succeed",
			members: []address{
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
				{
					Address: "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq",
					Choice:  commondao.ChoiceYes,
				},
			},
			success: true,
		},
		{
			name: "fail",
			members: []address{
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
				{
					Address: "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq",
					Choice:  commondao.ChoiceNo,
				},
			},
		},
		{
			name: "fail with no votes",
			members: []address{
				"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				"g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq",
			},
			err: commondao.ErrNoQuorum,
		},
		{
			name: "no quorum",
			members: []address{
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
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				p       dissolvePropDefinition
				record  commondao.VotingRecord
				members = commondao.NewMemberStorage()
			)

			for _, m := range tc.members {
				members.Add(m)
			}

			for _, v := range tc.votes {
				record.AddVote(v)
			}

			ctx := commondao.MustNewVotingContext(&record, members)

			success, err := p.Tally(ctx)

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

func TestDissolvePropDefinitionExecute(t *testing.T) {
	dao := commondao.New()
	urequire.False(t, dao.IsDeleted(), "expected DAO not to be soft deleted by default")

	p := newDissolvePropDefinition(dao)
	err := p.Execute(cross)
	urequire.NoError(t, err, "expected no execution error")

	uassert.True(t, dao.IsDeleted(), "expected DAO to be soft deleted")
}
