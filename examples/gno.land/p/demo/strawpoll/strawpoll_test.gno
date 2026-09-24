package strawpoll

import (
	"testing"

	"gno.land/p/samcrew/piechart"
)

func TestCreatePoll(t *testing.T) {
	t.Parallel()

	t.Run("EmptyQuestion", func(t *testing.T) {
		t.Parallel()

		emptyQuestion := ""

		_, err := New(emptyQuestion, []string{"Blue", "Green"})
		if err != ErrQuestionMustBeProvided {
			t.Errorf("expected error to be: %v, but got: %v", ErrQuestionMustBeProvided, err)
		}
	})

	t.Run("MissingOptions", func(t *testing.T) {
		t.Parallel()

		question := "What is your favorite color?"
		insufficientOptions := []string{"Blue"}

		_, err := New(question, insufficientOptions)
		if err != ErrMinimumOptionsMustBeProvided {
			t.Errorf("expected error to be: %v, but got: %v", ErrMinimumOptionsMustBeProvided, err)
		}
	})

	t.Run("TooManyOptions", func(t *testing.T) {
		t.Parallel()

		question := "What is your favorite color?"
		excessiveOptions := []string{"Blue", "Green", "Red", "Yellow", "Purple", "Orange", "Pink", "Brown", "Black", "White", "Gray"}

		_, err := New(question, excessiveOptions)
		if err != ErrMaximumOptionsExceeded {
			t.Errorf("expected error to be: %v, but got: %v", ErrMaximumOptionsExceeded, err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		question := "What is your favorite color?"
		options := []string{"Blue", "Green"}

		poll, err := New(question, options)
		if err != nil {
			t.Errorf("expected no error, but got: %v", err)
		}

		if poll.question != question {
			t.Errorf("expected question to be: %v, but got: %v", question, poll.question)
		}

		if !poll.votes.Has(options[0]) || !poll.votes.Has(options[1]) {
			t.Errorf("expected options to contain: %v and %v, but got: %v", options[0], options[1], poll.options)
		}

		for _, option := range options {
			if poll.VoteCount(option) != 0 {
				t.Errorf("expected votes for option %v to be 0, but got: %v", option, poll.VoteCount(option))
			}
		}
	})
}

func TestVote(t *testing.T) {
	t.Parallel()

	t.Run("AlreadyVoted", func(t *testing.T) {
		t.Parallel()

		question := "What is your favorite color?"
		options := []string{"Blue", "Green"}

		poll, err := New(question, options)
		if err != nil {
			t.Fatalf("failed to create valid poll: %v", err)
		}

		err = poll.Vote(options[0])
		if err != nil {
			t.Fatalf("failed to vote: %v", err)
		}

		err = poll.Vote(options[1])
		if err != ErrAlreadyVoted {
			t.Errorf("expected error to be: %v, but got: %v", ErrAlreadyVoted, err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		question := "What is your favorite color?"
		options := []string{"Blue", "Green"}

		poll, err := New(question, options)
		if err != nil {
			t.Fatalf("failed to create valid poll: %v", err)
		}

		err = poll.Vote(options[0])
		if err != nil {
			t.Fatalf("failed to vote: %v", err)
		}

		if poll.VoteCount(options[0]) != 1 {
			t.Errorf("expected votes for option %v to be 1, but got: %v", options[0], poll.VoteCount(options[0]))
		}
	})
}

func TestMarkdown(t *testing.T) {
	t.Parallel()

	question := "What is your favorite color?"
	options := []string{"Blue", "Green"}

	poll, err := New(question, options)
	if err != nil {
		t.Fatalf("failed to create valid poll: %v", err)
	}

	err = poll.Vote(options[0])
	if err != nil {
		t.Fatalf("failed to vote: %v", err)
	}

	want := piechart.Render([]piechart.PieSlice{
		{Value: 1, Label: options[0], Color: pieColors[0]},
		{Value: 0, Label: options[1], Color: pieColors[1]},
	}, question)

	if poll.Markdown() != want {
		t.Errorf("expected markdown to render following piechart: %q, but got: %q", want, poll.Markdown())
	}
}
