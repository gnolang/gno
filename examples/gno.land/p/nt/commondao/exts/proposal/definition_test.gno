package proposal_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"gno.land/p/nt/commondao"
	"gno.land/p/nt/commondao/exts/proposal"
	"gno.land/p/nt/uassert"
)

func TestNewDefinition(t *testing.T) {
	tests := []struct {
		name  string
		title string
		body  string
		err   error
	}{
		{
			name:  "ok",
			title: "Proposal Title",
			body:  "Foo",
		},
		{
			name:  "empty title",
			title: "",
			err:   proposal.ErrTitleIsRequired,
		},
		{
			name:  "empty title",
			title: "Proposal Title",
			body:  "",
			err:   proposal.ErrBodyIsRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			def, err := proposal.NewDefinition(tt.title, tt.body)

			// Assert
			if tt.err != nil {
				uassert.ErrorIs(t, err, tt.err, "expect an error")
				return
			}

			uassert.NoError(t, err, "expect no errors")
			uassert.Equal(t, tt.title, def.Title(), "expect title to match")
			uassert.Equal(t, tt.body, def.Body(), "expect body to match")
		})
	}
}

func TestDefinitionCustomVoteChoices(t *testing.T) {
	// Default voting choices
	def := proposal.MustNewDefinition("Title", "Body")
	uassert.Equal(t, "YES,NO,ABSTAIN", joinVoteChoices(def.CustomVoteChoices()), "expect default vote choces to match")

	// Custom voting choices
	def = proposal.MustNewDefinition("Title", "Body", proposal.WithVoteChoices([]commondao.VoteChoice{"A", "B", "C"}))
	uassert.Equal(t, "A,B,C", joinVoteChoices(def.CustomVoteChoices()), "expect custom vote choces to match")
}

func TestDefinitionVotingPeriod(t *testing.T) {
	// Default voting period
	def := proposal.MustNewDefinition("Title", "Body")
	uassert.True(t, proposal.DefaultVotingPeriod == def.VotingPeriod(), "expect default voting period to match")

	// Custom voting period
	period := time.Hour * 24 * 30
	def = proposal.MustNewDefinition("Title", "Body", proposal.WithVotingPeriod(period))
	uassert.True(t, period == def.VotingPeriod(), "expect custom voting period to match")

	// Zero voting period
	period = 0
	def = proposal.MustNewDefinition("Title", "Body", proposal.WithVotingPeriod(period))
	uassert.True(t, period == def.VotingPeriod(), "expect zero voting period to match")
}

func TestDefinitionTally(t *testing.T) {
	tests := []struct {
		name   string
		passes bool
		err    error
	}{
		{
			name:   "passes",
			passes: true,
		},
		{
			name:   "fails",
			passes: true,
		},
		{
			name: "error",
			err:  errors.New("no quorum"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			fn := func(commondao.VotingContext) (bool, error) {
				return tt.passes, tt.err
			}

			def := proposal.MustNewDefinition("Title", "Body", proposal.WithTally(fn))

			// Act
			passes, err := def.Tally(commondao.VotingContext{})

			// Assert
			if tt.err != nil {
				uassert.ErrorIs(t, err, tt.err, "expect tally to fail")
				return
			}

			uassert.NoError(t, err, "expect no errors")
			uassert.Equal(t, tt.passes, passes, "expect tally result to match")
		})
	}
}

func TestDefinitionValidation(t *testing.T) {
	errTest := errors.New("test error")
	tests := []struct {
		name      string
		validator func() error
		err       error
	}{
		{
			name:      "ok",
			validator: func() error { return nil },
		},
		{
			name:      "validation error",
			validator: func() error { return errTest },
			err:       errTest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			def := proposal.MustNewDefinition("Title", "Body", proposal.WithValidation(tt.validator))

			// Act
			err := def.Validate()

			// Assert
			if tt.err != nil {
				uassert.ErrorIs(t, err, tt.err, "expect validation to fail")
				return
			}

			uassert.NoError(t, err, "expect no errors")
		})
	}
}

func joinVoteChoices(choices []commondao.VoteChoice) string {
	values := make([]string, len(choices))
	for i, c := range choices {
		values[i] = string(c)
	}
	return strings.Join(values, ",")
}
