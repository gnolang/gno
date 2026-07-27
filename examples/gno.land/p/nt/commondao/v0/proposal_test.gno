package commondao_test

import (
	"errors"
	"testing"
	"time"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"

	"gno.land/p/nt/commondao/v0"
)

// choicesPropDef customizes the allowed vote choices.
type choicesPropDef struct {
	propDef
	choices []commondao.VoteChoice
}

func (d choicesPropDef) CustomVoteChoices() []commondao.VoteChoice { return d.choices }

// comboPropDef is invalid: it customizes vote choices while also being
// decided by the default YES/NO/ABSTAIN tally.
type comboPropDef struct{ choicesPropDef }

func (comboPropDef) Threshold() commondao.Threshold { return commondao.ThresholdSupermajority }

// validablePropDef supports state validation.
type validablePropDef struct {
	propDef
	err error
}

func (d validablePropDef) Validate() error { return d.err }

func TestProposalNew(t *testing.T) {
	t.Run("definition is required", func(t *testing.T) {
		_, err := commondao.NewProposal(1, memberA, nil)
		uassert.ErrorIs(t, err, commondao.ErrProposalDefinitionRequired)
	})

	t.Run("creator must be valid", func(t *testing.T) {
		_, err := commondao.NewProposal(1, address(""), propDef{})
		uassert.ErrorIs(t, err, commondao.ErrInvalidCreatorAddress)
	})

	t.Run("defaults", func(t *testing.T) {
		p, err := commondao.NewProposal(42, memberA, propDef{title: "T", body: "B", period: time.Hour})
		urequire.NoError(t, err)

		uassert.Equal(t, uint64(42), p.ID())
		uassert.Equal(t, string(commondao.StatusActive), string(p.Status()))
		uassert.Equal(t, memberA, p.Creator())
		uassert.Equal(t, "T", p.Definition().Title())
		uassert.Equal(t, "B", p.Definition().Body())
		uassert.Equal(t, 0, p.VotingRecord().Size())
		uassert.Equal(t, "", p.StatusReason())
		uassert.False(t, p.HasVotingDeadlinePassed())

		// Proposals created outside CommonDAO.Propose have an empty
		// electorate and therefore reject every vote.
		uassert.Equal(t, 0, p.Electorate().Size())

		choices := p.VoteChoices()
		urequire.Equal(t, 3, len(choices), "expect default vote choices")
		uassert.True(t, p.IsVoteChoiceValid(commondao.ChoiceYes))
		uassert.True(t, p.IsVoteChoiceValid(commondao.ChoiceNo))
		uassert.True(t, p.IsVoteChoiceValid(commondao.ChoiceAbstain))
		uassert.False(t, p.IsVoteChoiceValid(commondao.VoteChoice("FOO")))
	})

	t.Run("custom vote choices", func(t *testing.T) {
		d := choicesPropDef{choices: []commondao.VoteChoice{"A", "B"}}
		p, err := commondao.NewProposal(1, memberA, d)
		urequire.NoError(t, err)

		uassert.Equal(t, 2, len(p.VoteChoices()))
		uassert.True(t, p.IsVoteChoiceValid(commondao.VoteChoice("A")))
		uassert.False(t, p.IsVoteChoiceValid(commondao.ChoiceYes))
	})

	t.Run("single custom choice falls back to defaults", func(t *testing.T) {
		d := choicesPropDef{choices: []commondao.VoteChoice{"A"}}
		p, err := commondao.NewProposal(1, memberA, d)
		urequire.NoError(t, err)

		uassert.True(t, p.IsVoteChoiceValid(commondao.ChoiceYes))
		uassert.True(t, p.IsVoteChoiceValid(commondao.ChoiceNo))
		uassert.True(t, p.IsVoteChoiceValid(commondao.ChoiceAbstain))
	})

	t.Run("too many custom choices", func(t *testing.T) {
		choices := make([]commondao.VoteChoice, commondao.MaxCustomVoteChoices+1)
		for i := range choices {
			choices[i] = commondao.VoteChoice(string(rune('A' + i)))
		}

		_, err := commondao.NewProposal(1, memberA, choicesPropDef{choices: choices})
		uassert.ErrorIs(t, err, commondao.ErrMaxCustomVoteChoices)
	})

	t.Run("default tallier cannot customize choices", func(t *testing.T) {
		_, err := commondao.NewProposal(1, memberA, comboPropDef{})
		uassert.ErrorIs(t, err, commondao.ErrCustomChoicesNotAllowed)
	})
}

func TestProposalValidate(t *testing.T) {
	t.Run("non validable definitions are always valid", func(t *testing.T) {
		p, err := commondao.NewProposal(1, memberA, propDef{})
		urequire.NoError(t, err)
		uassert.NoError(t, p.Validate())
	})

	t.Run("validable definitions validate while executable", func(t *testing.T) {
		boom := errors.New("boom")
		p, err := commondao.NewProposal(1, memberA, validablePropDef{err: boom})
		urequire.NoError(t, err)
		uassert.ErrorIs(t, p.Validate(), boom)
	})
}

func TestMustValidate(cur realm, t *testing.T) {
	urequire.NotPanics(t, cur, func() {
		commondao.MustValidate(validablePropDef{})
	})

	urequire.PanicsWithMessage(t, cur, "boom", func() {
		commondao.MustValidate(validablePropDef{err: errors.New("boom")})
	})
}
