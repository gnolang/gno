package commondao

import (
	"std"
	"strings"
	"testing"
	"time"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	"gno.land/p/nt/commondao"
)

var _ commondao.ProposalDefinition = (*TextPropDefinition)(nil)

func TestTextPropDefinitionNew(t *testing.T) {
	cases := []struct {
		name         string
		title, body  string
		quorum       float64
		votingPeriod time.Duration
		panicMsg     string
	}{
		{
			name:         "ok",
			title:        "Test",
			body:         "Foo",
			votingPeriod: time.Hour * 24 * 7,
			quorum:       commondao.QuorumOneThird,
		},
		{
			name:     "empty proposal title",
			title:    "",
			panicMsg: "proposal title is empty",
		},
		{
			name:     "long proposal title",
			title:    strings.Repeat("A", 256),
			panicMsg: "proposal title is too long, max length is 255 chars",
		},
		{
			name:     "empty proposal body",
			title:    "Test",
			body:     "",
			panicMsg: "proposal body is empty",
		},
		{
			name:     "long proposal body",
			title:    "Test",
			body:     strings.Repeat("A", 15001),
			panicMsg: "proposal body is too long, max length is 15000 chars",
		},
		{
			name:         "invalid voting period",
			title:        "Test",
			body:         "Foo",
			votingPeriod: time.Hour * 10,
			panicMsg:     "minimum proposal voting period is one day",
		},
		{
			name:         "invalid quorum",
			title:        "Test",
			body:         "Foo",
			votingPeriod: time.Hour * 24 * 7,
			quorum:       0.32,
			panicMsg:     "minimum quorum allowed is one third (0.33)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.panicMsg != "" {
				urequire.PanicsWithMessage(t, tc.panicMsg, func() {
					NewTextPropDefinition(tc.title, tc.body, tc.quorum, tc.votingPeriod)
				}, "expected definition to panic with message")

				return
			}

			urequire.NotPanics(t, func() {
				NewTextPropDefinition(tc.title, tc.body, tc.quorum, tc.votingPeriod)
			}, "expected definition to be created")
		})
	}
}

func TestTextPropDefinitionTally(t *testing.T) {
	cases := []struct {
		name    string
		members []std.Address
		votes   []commondao.Vote
		err     error
		success bool
	}{
		{
			name: "succeed",
			members: []std.Address{
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
			members: []std.Address{
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
			members: []std.Address{
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
				record  commondao.VotingRecord
				p       = TextPropDefinition{quorum: commondao.QuorumTwoThirds}
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
