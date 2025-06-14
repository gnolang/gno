package poll

import (
	"std"
	"testing"
	"time"
)

func TestNewPoll(t *testing.T) {
	creator := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	now := time.Now()
	deadline := now.Add(time.Hour * 24) // 1 day from now

	p := NewPoll(
		"Favorite Color",
		"What is your favorite color?",
		[]string{"Red", "Blue", "Green"},
		1,
		deadline,
		creator,
	)

	if p.Title != "Favorite Color" {
		t.Errorf("expected title 'Favorite Color', got %s", p.Title)
	}
	if len(p.Options) != 3 {
		t.Errorf("expected 3 options, got %d", len(p.Options))
	}
	if p.MaxChoices != 1 {
		t.Errorf("expected maxChoices 1, got %d", p.MaxChoices)
	}
	if p.Creator != creator {
		t.Errorf("expected creator %s, got %s", creator, p.Creator)
	}
}

func TestNewPollValidation(t *testing.T) {
	creator := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	now := time.Now()
	pastDeadline := now.Add(-time.Hour) // 1 hour ago

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for past deadline")
		}
	}()

	NewPoll(
		"Test Poll",
		"Description",
		[]string{},
		1,
		pastDeadline,
		creator,
	)
}

func TestVoting(t *testing.T) {
	creator := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	voter := std.Address("g1htpxzv2dkplvzg50nd8fuwt5yg4jn8dgtx076k")
	now := time.Now()
	deadline := now.Add(time.Hour * 24)

	p := NewPoll(
		"Test Poll",
		"Description",
		[]string{"Option 1", "Option 2", "Option 3"},
		2, // max two choices
		deadline,
		creator,
	)

	// Test single vote
	p.Vote(voter, []int{1})

	if !p.HasVoted(voter) {
		t.Error("expected HasVoted to return true")
	}

	tally := p.Tally()
	if tally[1] != 1 {
		t.Errorf("expected tally[1] = 1, got %d", tally[1])
	}

	// Test double voting (should panic)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on double vote")
		}
	}()
	p.Vote(voter, []int{0})
}

func TestOptionManagement(t *testing.T) {
	creator := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	now := time.Now()
	deadline := now.Add(time.Hour * 24)

	p := NewPoll(
		"Test Poll",
		"Description",
		[]string{},
		1,
		deadline,
		creator,
	)

	// Test adding options
	p.AddOption("Option 1")
	p.AddOption("Option 2")

	if len(p.Options) != 2 {
		t.Errorf("expected 2 options, got %d", len(p.Options))
	}

	// Test editing option
	p.EditOption(0, "Updated Option 1")
	if p.Options[0] != "Updated Option 1" {
		t.Errorf("expected 'Updated Option 1', got %s", p.Options[0])
	}

	// Test deleting option
	p.DeleteOption(0)
	if len(p.Options) != 1 {
		t.Errorf("expected 1 option after deletion, got %d", len(p.Options))
	}
	if p.Options[0] != "Option 2" {
		t.Errorf("expected remaining option to be 'Option 2', got %s", p.Options[0])
	}
}

func TestPollEdit(t *testing.T) {
	creator := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	nonCreator := std.Address("g1htpxzv2dkplvzg50nd8fuwt5yg4jn8dgtx076k")
	now := time.Now()
	deadline := now.Add(time.Hour * 24)
	newDeadline := now.Add(time.Hour * 48)

	p := NewPoll(
		"Original Title",
		"Original Description",
		[]string{"Option 1"},
		1,
		deadline,
		creator,
	)

	// Test successful edit
	p.Edit(
		creator,
		"New Title",
		"New Description",
		newDeadline,
		2,
	)

	if p.Title != "New Title" {
		t.Errorf("expected title 'New Title', got %s", p.Title)
	}
	if p.MaxChoices != 2 {
		t.Errorf("expected maxChoices 2, got %d", p.MaxChoices)
	}

	// Test edit by non-creator (should panic)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when non-creator tries to edit")
		}
	}()
	p.Edit(
		nonCreator,
		"Unauthorized Edit",
		"Should Fail",
		newDeadline,
		1,
	)
}

func TestVoterCount(t *testing.T) {
	creator := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	voter1 := std.Address("g1htpxzv2dkplvzg50nd8fuwt5yg4jn8dgtx076k")
	voter2 := std.Address("g1w3jhxarpv3m8ddxhqj2es8wzcgv4gzurvqc843")
	now := time.Now()
	deadline := now.Add(time.Hour * 24)

	p := NewPoll(
		"Test Poll",
		"Description",
		[]string{"Option 1", "Option 2", "Option 3"},
		2,
		deadline,
		creator,
	)

	if count := p.VoterCount(); count != 0 {
		t.Errorf("expected initial voter count 0, got %d", count)
	}

	p.Vote(voter1, []int{0})
	if count := p.VoterCount(); count != 1 {
		t.Errorf("expected voter count 1, got %d", count)
	}

	p.Vote(voter2, []int{1, 2})
	if count := p.VoterCount(); count != 2 {
		t.Errorf("expected voter count 2, got %d", count)
	}
}

func TestOptionManagementAfterVoting(t *testing.T) {
	creator := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	voter := std.Address("g1htpxzv2dkplvzg50nd8fuwt5yg4jn8dgtx076k")
	now := time.Now()
	deadline := now.Add(time.Hour * 24)

	p := NewPoll(
		"Test Poll",
		"Description",
		[]string{"Option 1", "Option 2"},
		1,
		deadline,
		creator,
	)

	// Cast a vote
	p.Vote(voter, []int{0})

	// Test adding option after voting
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when adding option after voting")
		}
	}()
	p.AddOption("Option 3")

	// Test editing option after voting
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when editing option after voting")
		}
	}()
	p.EditOption(0, "Updated Option 1")

	// Test deleting option after voting
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when deleting option after voting")
		}
	}()
	p.DeleteOption(0)
}
