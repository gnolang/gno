package strawpoll

import (
	"strings"
	"testing"

	"gno.land/p/demo/strawpoll"
)

func TestCreatePoll(t *testing.T) {
	t.Run("MissingQuestion", func(t *testing.T) {
		emptyQuestion := ""

		err := CreatePoll(cross, emptyQuestion, "Blue,Green")
		if err != strawpoll.ErrQuestionMustBeProvided {
			t.Errorf("expected error to be: %v, but got: %v", strawpoll.ErrQuestionMustBeProvided, err)
		}
	})

	t.Run("MissingOptions", func(t *testing.T) {
		question := "What is your favorite color?"
		insufficientOptions := "Blue"

		err := CreatePoll(cross, question, insufficientOptions)
		if err != strawpoll.ErrMinimumOptionsMustBeProvided {
			t.Errorf("expected error to be: %v, but got: %v", strawpoll.ErrMinimumOptionsMustBeProvided, err)
		}
	})

	t.Run("TooManyOptions", func(t *testing.T) {
		question := "What is your favorite color?"
		excessiveOptions := "Blue,Green,Red,Yellow,Purple,Orange,Pink,Brown,Black,White,Gray"

		err := CreatePoll(cross, question, excessiveOptions)
		if err != strawpoll.ErrMaximumOptionsExceeded {
			t.Errorf("expected error to be: %v, but got: %v", strawpoll.ErrMaximumOptionsExceeded, err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		question := "What is your favorite color?"
		options := []string{"Blue", "Green"}

		err := CreatePoll(cross, question, strings.Join(options, ","))
		if err != nil {
			t.Errorf("expected no error, but got: %v", err)
		}

		if len(polls) != 1 {
			t.Errorf("expected 1 poll to be created, but got: %v", len(polls))
		}
	})
}

func TestVote(t *testing.T) {
	t.Run("PollNotFound", func(t *testing.T) {
		err := Vote(cross, 999, "Blue")
		if err != ErrPollNotFound {
			t.Errorf("expected error to be: %v, but got: %v", ErrPollNotFound, err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		question := "What is your favorite color?"
		options := []string{"Blue", "Green"}

		err := CreatePoll(cross, question, strings.Join(options, ","))
		if err != nil {
			t.Fatalf("failed to create poll: %v", err)
		}

		err = Vote(cross, int64(len(polls)), options[0])
		if err != nil {
			t.Fatalf("failed to vote: %v", err)
		}

		lastPoll := polls[len(polls)-1]

		if lastPoll.VoteCount(options[0]) != 1 {
			t.Errorf("expected votes for option %v to be 1, but got: %v", options[0], lastPoll.VoteCount(options[0]))
		}
	})
}
