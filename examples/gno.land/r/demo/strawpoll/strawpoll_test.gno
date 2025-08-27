package strawpoll

import (
	"strings"
	"testing"
)

func TestCreatePoll(t *testing.T) {
	t.Run("MissingQuestion", func(t *testing.T) {
		emptyQuestion := ""

		err := CreatePoll(cross, emptyQuestion, "Blue,Green")
		if err != ErrQuestionMustBeProvided {
			t.Errorf("expected error to be: %v, but got: %v", ErrQuestionMustBeProvided, err)
		}
	})

	t.Run("MissingOptions", func(t *testing.T) {
		question := "What is your favorite color?"
		insufficientOptions := "Blue"

		err := CreatePoll(cross, question, insufficientOptions)
		if err != ErrMiniminalOptionsMustBeProvided {
			t.Errorf("expected error to be: %v, but got: %v", ErrMiniminalOptionsMustBeProvided, err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		question := "What is your favorite color?"
		options := []string{"Blue", "Green"}

		err := CreatePoll(cross, question, strings.Join(options, ","))
		if err != nil {
			t.Errorf("expected no error, but got: %v", err)
		}

		lastPoll := polls[len(polls)-1]

		if lastPoll.Question != question {
			t.Errorf("expected question to be: %v, but got: %v", question, lastPoll.Question)
		}

		if len(lastPoll.Options) != 2 {
			t.Errorf("expected options length to be: %v, but got: %v", len(options), len(lastPoll.Options))
		}

		for i, option := range options {
			if lastPoll.Options[i] != option {
				t.Errorf("expected option at index %d to be: %v, but got: %v", i, option, lastPoll.Options[i])
			}

			if lastPoll.Votes[option] != 0 {
				t.Errorf("expected votes for option %v to be 0, but got: %v", option, lastPoll.Votes[option])
			}
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

	t.Run("AlreadyVoted", func(t *testing.T) {
		question := "What is your favorite color?"
		options := []string{"Blue", "Green"}

		err := CreatePoll(cross, question, strings.Join(options, ","))
		if err != nil {
			t.Fatalf("failed to create poll: %v", err)
		}

		pollNumber := int64(len(polls))

		err = Vote(cross, pollNumber, options[0])
		if err != nil {
			t.Fatalf("failed to vote: %v", err)
		}

		err = Vote(cross, pollNumber, options[1])
		if err != ErrAlreadyVoted {
			t.Errorf("expected error to be: %v, but got: %v", ErrAlreadyVoted, err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		question := "What is your favorite color?"
		options := []string{"Blue", "Green"}

		err := CreatePoll(cross, question, strings.Join(options, ","))
		if err != nil {
			t.Fatalf("failed to create poll: %v", err)
		}

		pollNumber := len(polls)

		if pollNumber < 1 {
			t.Fatalf("no polls available to vote on")
		}

		err = Vote(cross, int64(pollNumber), options[0])
		if err != nil {
			t.Fatalf("failed to vote: %v", err)
		}

		poll := polls[pollNumber-1]

		if poll.Votes[options[0]] != 1 {
			t.Errorf("expected votes for option %v to be 1, but got: %v", options[0], poll.Votes[options[0]])
		}
	})
}
