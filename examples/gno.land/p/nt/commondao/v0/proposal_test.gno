package commondao_test

import (
	"errors"
	"testing"
	"time"

	"gno.land/p/nt/uassert/v0"

	"gno.land/p/nt/commondao/v0"
)

// validablePropDef supports state validation.
type validablePropDef struct {
	propDef
	err error
}

func (d validablePropDef) Validate() error { return d.err }

func TestProposalNew(t *testing.T) {
	dao := commondao.New(commondao.WithCouncilMember(memberA))

	t.Run("creator must be valid", func(t *testing.T) {
		_, err := dao.Propose(address(""), propDef{})
		uassert.ErrorIs(t, err, commondao.ErrInvalidCreatorAddress)
	})

	t.Run("defaults", func(t *testing.T) {
		p := mustPropose(t, dao, memberA, propDef{title: "T", body: "B", period: time.Hour})

		uassert.Equal(t, string(commondao.StatusActive), string(p.Status()))
		uassert.Equal(t, memberA, p.Creator())
		uassert.Equal(t, "T", p.Definition().Title())
		uassert.Equal(t, "B", p.Definition().Body())
		uassert.Equal(t, 0, p.VotingRecord().Size())
		uassert.Equal(t, "", p.StatusReason())
		uassert.False(t, p.HasVotingDeadlinePassed())

		// The vote choices are fixed by the default voting rules
		uassert.True(t, p.IsVoteChoiceValid(commondao.ChoiceYes))
		uassert.True(t, p.IsVoteChoiceValid(commondao.ChoiceNo))
		uassert.True(t, p.IsVoteChoiceValid(commondao.ChoiceAbstain))
		uassert.False(t, p.IsVoteChoiceValid(commondao.VoteChoice("FOO")))
	})
}

func TestProposalValidate(t *testing.T) {
	dao := commondao.New(commondao.WithCouncilMember(memberA))

	t.Run("non validable definitions are always valid", func(t *testing.T) {
		p := mustPropose(t, dao, memberA, propDef{period: time.Hour})
		uassert.NoError(t, p.Validate())
	})

	t.Run("validable definitions validate while executable", func(t *testing.T) {
		boom := errors.New("boom")
		p := mustPropose(t, dao, memberA, validablePropDef{
			propDef: propDef{period: time.Hour},
			err:     boom,
		})
		uassert.ErrorIs(t, p.Validate(), boom)
	})
}
